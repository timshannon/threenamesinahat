# Three Names in a Hat

The party game *Three Names in a Hat* also known as *Celebrity* in some circles played in the format
of an online game where players join and play on their own phones / tablets / computers.


## Motivation
This simple party game is one my friends and family always enjoy playing.  However the recent Covid-19 world-wide
pandemic has made playing party games a little difficult. I felt it was the perfect quarentine project to turn this
simple party game into one that could be played remotely while we all are social distancing.

Around the world right now, musicians are playing from balconies, writers are furiously writing blog posts and tweets,
and artists and sharing their works over instagram. I wanted to make someting small and fun to mark this historic
time, and I really have only one skill to offer.

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

**Round 2**: Same as previous rounds, except no words, only silent acting out clues.

**Round 3**: Same as round one except this time you only get one word.

After three rounds, the team with the most points win.


## Overview of play

1. One player will start a game. 
    * The game will have a randomly generated code
2. Other players can join the game using the game code
3. The first two players in get to choose team names and team colors
    * Defaults to Team Blue and Team Red
4. More players can continue to join, and pick which team they want to join
5. Game creator (1st player in) gets to choose when to start the game
6. When the game starts, all players will have 1-2 minutes to write down 3 names
7. A team is randomly selected to start, and player order is randomly chosen at this time
8. First player is given a start button
9. Player presses start and starts the timer and is given the first randomly selected name
10. Player presses next when the name is guessed, and they get a point for each time they press next before the
timer runs out
11. Point totals accumlate on everyone's screen as game progresses
12. When timer runs out, next player on next team gets a start button
13. Play continues until all names are chosen and all rounds are played.
14. Game summary and winning team is shown as well as stats such as who's turn got the most guesses
15. Option to replay with same team start a new game


## Goals
* Simple
* No database, everything is tracked in memory
* No logons, no sessions, no cookies
* No SSL, run it behind nginx / traefik / apache / etc and let them handle ssl
* No configuration
* One command line flag *port*
* Built *mostly* with the core Go libraries