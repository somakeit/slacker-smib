[![Build Status](https://travis-ci.org/somakeit/slacker-smib.svg?branch=master)](https://travis-ci.org/somakeit/slacker-smib)
[![Coverage Status](https://coveralls.io/repos/github/somakeit/slacker-smib/badge.svg)](https://coveralls.io/github/somakeit/slacker-smib)

slacker-smib
============
A smib which is even more slack, this bot runs in the So Make It slack instance.

The master branch of this repo is automatically deployed to toad and started.

Commands
--------
All the commands run by this bot are in the repo [smib-commands](https://github.com/somakeit/smib-commands). They can be written in any language. The arguments are compatible with [smib](https://github.com/somakeit/smib) (the IRC bot) and are as so:
 * $1 - User, the user calling the script, this is is a slack syntax for mentioning the user, it does not look like the user's name.
 * $2 - Channel, the channel the command was invoked from, or "null" if it was not a channel. This is the display name of the channel.
 * $3 - Sender, the Channel for channel message or the User if it was not a channel message. This is a legacy argument.
 * $4 - Args, everything the user said after the command and one space.
 * $5 - Command, what the user typed to get this command, will differ from $0, may be a prefix of the full command.
 * $6 - UserDisplay, the display name of the user. The IRC bot does not send this.