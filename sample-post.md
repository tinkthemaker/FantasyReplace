---
title: Building a keyboard macro pad
date: 2026-06-08
tags: [hardware, build-notes]
---

# Building a keyboard macro pad

Today I finally finished my macro pad build. The hardware was easy, but the firmware was really annoying.

I started by soldering the switches to the PCB. My soldering is bad, so I expected problems. The first test failed with an error: the third column was dead. I spent an hour debugging before I found a broken solder joint.

For the firmware I used [QMK](https://qmk.fm), which is a great framework. You just edit a config file and compile:

```bash
qmk compile -kb mypad -km default
qmk flash
```

The `keymap.c` file holds the layout. I committed everything to my repo so I have a backup.

Things I learned:

- Check your solder joints before testing
- The compiler errors are weird but the logs help
- Maybe buy a better soldering iron, the cheap one made this harder

Next I want to deploy a script that syncs my config to the cloud automatically. Probably next week.
