---
title: "tinksoft.com is live"
date: 2026-06-10
description: "A build log for the work — and the work it now tracks."
tags: [meta, tinksoft-site]
---

I build a lot of things. Lately I've been finishing more of them, and I
wanted a place to write that down honestly. This is that place.

## why a build log

Most project pages I see are sales decks. The interesting part is the
middle — the bug that ate a Saturday, the rewrite I should have done
three months earlier, the moment the architecture actually clicked.
I wanted a place to put *that*, not the polished "look what I shipped"
recap. So this is it.

The constraint is the point: every page is a single HTTP request,
ships zero JavaScript, and fits in a couple of KB. Constraints aren't
the obstacle, they're the design.

## what you'll find here

There's a [build log](/) for the running commentary and a
[projects page](/projects) for the work itself — the things I've
shipped, the things on the workbench, and the things still in the
sketchbook. Each project has its own page with the full story and
every related log entry.

Right now that means:

- **[SepulchrynScan](/projects/sepulchrynscan/)** — a multithreaded
  service-discovery and port scanner in Go. The one I reach for first
  when I need to map a network.
- **[Detectsmith](/projects/detectsmith/)** — a detection engineering
  workbench that tries to be useful, not Splunk.
- **[CyberToolbox](/projects/cybertoolbox/)** — a curated CLI bundle
  for the security tasks I do often enough to want a single binary for.
- **[Locket](/projects/locket/)** — an encrypted key vault TUI. I got
  tired of `pass` syntax and wanted something I'd enjoy typing into.
- **[Shipnote](/projects/shipnote/)** — release notes that read my git
  log and write a changelog I don't have to.
- **[Cleanpaste](/projects/cleanpaste/)** — a clipboard sanitizer.
  Strips invisible characters, normalizes whitespace, kills tracking
  pixels. Solves a real problem I keep seeing.
- **[Sigil](/projects/sigil/)** — a tiny CLI for signing and
  verifying release artifacts. I wanted GPG to be less awful.
- **[TPing](/projects/tping/)** — a Windows on-call network monitor
  for store devices. A TUI for pinging everything and seeing the
  picture at a glance.
- **[PulpDescent](/projects/pulpdescent/)** and
  **[aGrandAdventure](/projects/agrandadventure/)** — two LLM-less
  procedural text dungeons in Godot 4. The constraint of *no LLM*
  turned out to be the most interesting one I've ever worked under.

## what happens next

New posts go up as the work does — usually when something interesting
breaks, gets fixed, or gets shipped. There's an [RSS feed](/rss.xml)
if that's how you read.

If you want to follow along, the feed is the easiest way. Otherwise,
just check the [build log](/) when you remember.
