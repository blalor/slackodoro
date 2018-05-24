package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"gopkg.in/ini.v1"

	"github.com/nlopes/slack"
)

var opts struct {
	Profiles []string `long:"profile"`
	IniFile  string   `long:"config" default:"~/.slackodoro/config"`

	Command struct {
		Argv0 string
		Argv  []string
	} `positional-args:"yes" required:"yes"`
}

type Slacker struct {
	team   string
	token  string
	userId string

	originalStatus struct {
		emoji string
		text  string
	}

	originalSnoozeEnabled bool

	client *slack.Client
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.IniFile[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("unable to retrieve current user: %s", err)
		}

		opts.IniFile = filepath.Join(usr.HomeDir, opts.IniFile[2:])
	}

	cfg, err := ini.Load(opts.IniFile)
	if err != nil {
		log.Fatalf("unable to load %s: %s", opts.IniFile, err)
	}

	if len(opts.Profiles) == 0 {
		defaultProfile := cfg.Section("").Key("default").String()
		if defaultProfile == "" {
			log.Fatal("no profiles provided and no 'default' key found")
		}

		opts.Profiles = []string{defaultProfile}
	}

	// initialize and record all necessary info
	slackers := make([]*Slacker, len(opts.Profiles))
	for i, profile := range opts.Profiles {
		slacker := &Slacker{
			team: profile,
		}

		sec, err := cfg.GetSection(profile)
		if err != nil {
			log.Fatalf("no section %s", profile)
		}

		key, err := sec.GetKey("token")
		if err != nil {
			log.Fatalf("no 'token' key in section %s", profile)
		}

		slacker.token = key.String()
		if slacker.token == "" {
			log.Fatalf("token for %s is empty", slacker.team)
		}

		slacker.client = slack.New(slacker.token)
		resp, err := slacker.client.AuthTest()
		if err != nil {
			log.Fatalf("slack client auth failed for %s: %s", slacker.team, err)
		}

		slacker.userId = resp.UserID

		ui, err := slacker.client.GetUserInfo(slacker.userId)
		if err != nil {
			log.Fatalf("unable to retrieve user info: %s", err)
		}

		slacker.originalStatus.emoji = ui.Profile.StatusEmoji
		slacker.originalStatus.text = ui.Profile.StatusText

		dnd, err := slacker.client.GetDNDInfo(&slacker.userId)
		if err != nil {
			log.Fatalf("unable to retrieve dnd info: %s", err)
		}

		slacker.originalSnoozeEnabled = dnd.SnoozeEnabled

		slackers[i] = slacker
	}

	// prepare
	for _, slacker := range slackers {
		log.Printf("shushing %s", slacker.team)

		if !slacker.originalSnoozeEnabled {
			snoozeDuration := 60 * 24 // 24h

			if _, err := slacker.client.SetSnooze(snoozeDuration); err != nil {
				log.Printf("unable to snooze %s: %v", slacker.team, err)
			}
		}

		if err := slacker.client.SetUserCustomStatus("slackodoro", "🍅"); err != nil {
			log.Printf("unable to set status for %s: %v", slacker.team, err)
		}
	}

	defer func() {
		// bring tha noize everything
		for _, slacker := range slackers {
			log.Printf("resetting %s", slacker.team)

			if err := slacker.client.SetUserCustomStatus(slacker.originalStatus.text, slacker.originalStatus.emoji); err != nil {
				log.Printf("unable to reset status for %s: %v", slacker.team, err)
			}

			if !slacker.originalSnoozeEnabled {
				if _, err := slacker.client.EndSnooze(); err != nil {
					log.Printf("unable to end DND/snooze for %s: %v", slacker.team, err)
				}
			}
		}
	}()

	// run command
    cmd := exec.Command(opts.Command.Argv0, opts.Command.Argv...)
    cmd.Stdin = nil
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    log.Printf("running %s %v", opts.Command.Argv0, opts.Command.Argv)
	if err = cmd.Run(); err != nil {
		log.Fatalf("error running %s %v: %v", opts.Command.Argv0, opts.Command.Argv, err)
	}
}
