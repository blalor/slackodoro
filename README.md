let your slack teammates know you're busy

# 🍅

This is a helper tool for shushing notifications from Slack while doing a pomodoro.  It snoozes your notifications, updates your status, and then resets both to their old values upon completion of the child process.

# usage

## config

Get a [legacy token](https://api.slack.com/custom-integrations/legacy-tokens) from Slack for each workspace you use.  Create `~/.slackodoro/config` like so:

    default = "my_workspace"
    
    [my_workspace]
    token = "xoxp-…"

Create a section for every team/workspace you would like to affect.  The `default` workspace name is optional.

## running

    slackodoro -- sleep $(( 25 * 60 ))

If you want to mute multiple profiles at once:

    slackodoro --profile team1 --profile team2 -- sleep $(( 25 * 60 ))

Or, if you're using [openpomodoro-cli](https://github.com/open-pomodoro/openpomodoro-cli/), which doesn't ([yet!](https://github.com/open-pomodoro/openpomodoro-cli/issues/1)) have native Slack integration:

    slackodoro -- pomodoro start --wait

## building

You need an existing Go development environment.

    go get github.com/blalor/slackodoro

Put the resulting `slackodoro` binary into your path.  If you're not sure where that might be, because Go is inscrutible, it's probably in `$( go env GOPATH )/bin`.
