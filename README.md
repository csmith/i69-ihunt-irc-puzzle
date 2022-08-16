# i69 IRC iHunt Puzzle

This repository contains the supporting materials for the puzzle
"Now you're gone" which was part of the [iHunt](https://theihunt.uk)
at i69.

## Ergo

The `ergo` directory contains a config file for the [ergo](https://ergo.chat/about)
ircd which:

- locks down account and channel creation
- disables most modern features
- defines two ircop accounts:
  - `anna` for the bot, with password `*Qc$ZRDB8WT2Kw`
  - `admin` for the admin staff, with password `dWbv^4xMZGRfqJ`
- defaults users to +R so they can't receive private messages
- enables IP masking

It also contains a simple MOTD file.

## Anna

This is a bot written in Go using the [ircevent](https://pkg.go.dev/github.com/ergochat/irc-go/ircevent)
package. It connects, opers up, and then listens for events by monitoring the
server notices (SNOTICES) sent by ergo.

For each user that connects, Anna creates a randomly named channel and force-joins
them to it. The channel topic instructs people to "chat but obey all the rules".

When a user speaks in a channel, their message is evaluated against a set of around
10 rules. If the user breaks any of the rules they are kicked from the channel with
the description of the rule that they broke. They can then rejoin and try speaking
again.

The rules leave only one specific word available for users to say in the channel,
and this is the answer to the puzzle.

Anna reports each message to a private #admin channel which is only joinable by
ircops, so the admin team can monitor progress. 

Finally, Anna monitors for users that change nicks and disconnect so she can tidy
up channels once the original user left.

## Docker files

The docker-compose file at the top level is enough to bring the entire stack (ergo
ircd and the Anna bot) up with a `docker compose up -d`. If changes are subsequently
made to Anna then you'll need to rebuild the image with `docker compose build`.

----

To the extent possible under law, I waive all copyright and related or neighbouring rights to this work. This work is
published from the United Kingdom. See [LICENCE.md](LICENCE.md) for full details.
