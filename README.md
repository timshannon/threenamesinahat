# Three Names in a Hat

The party game *Three Names in a Hat* also known as *Celebrity* in some circles played in the format
of an online game where players join and play on their own phones / tablets / computers.

## Quickstart

You can play it at: [https://www.threenamesinahat.com](https://www.threenamesinahat.com)

Or grab the code and run it yourself.

```
git clone git@github.com:timshannon/threenamesinahat.git
go build
./threenamesinahat
```

Or with docker

```
docker build --tag threenamesinahat https://github.com/timshannon/threenamesinahat.git#master
docker run --publish 8080:8080 threenamesinahat
```

## Motivation
This simple party game is one my friends and family always enjoy playing.  However the recent Covid-19 world-wide
pandemic has made playing party games a little difficult. I felt it was the perfect quarantine project to turn this
simple party game into one that could be played remotely while we all are social distancing.

Around the world right now, musicians are playing from balconies, writers are furiously writing blog posts and tweets,
and artists and sharing their works over instagram. I wanted to make something small and fun to mark this 
*strange and historic* time, and I really have only one skill to offer.

## Rules

For those who don't know the rules, or perhaps play by a different set of rules, here are the rules I'm building off of.

The players are split into two teams.  Each player writes down three names, and places them into the *hat*. The game then
proceeds with each team taking turns.  On each team's turn a player from that team has 30 seconds to try to get their team
to guess as many names as possible, pulling a new name from the hat when one is guessed.  After 30 seconds is up, if the
current name hasn't been guessed, the competing team gets a chance to *steal* that name before their round beings.

Players on each team all take turns each round, so everyone gets a chance to try to get their team to guess.

The game is broken into three rounds. Each round goes until there are no more names left in the hat, and the next round
starts with all the names put back in.

**Round 1**: The person who is up has to get their team to guess the current name by saying anything as long as it's
not part of the person's name or a direct reference to their name. For example they can't do things like "rhymes with"
or "starts with X" or "ends with X".

**Round 2**: Same as previous rounds, except no words, only acting out silent clues.

**Round 3**: Same as round one except this time you only get one word.

After three rounds, the team with the most points win.


## Goals
* Simple
* No database, everything is tracked in memory
* No logons, no sessions, no cookies
* No SSL, run it behind nginx / traefik / apache / etc and let them handle ssl
* No configuration
* One command line flag *port*
* Built *mostly* with the core Go libraries
* No NPM or transpiling
* No IE11 support