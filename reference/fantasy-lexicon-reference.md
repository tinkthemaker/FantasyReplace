# Fantasy/Wizard Lexicon Reference

Source material for expanding `lexicon.json`. Drawn from Warhammer Fantasy
Battles, Dungeons & Dragons, Terry Pratchett's Discworld, and World of Warcraft.

## Warhammer Fantasy — Colleges of Magic

The eight Winds of Magic, each with a corresponding College, Lore, and color.
Useful for tier3 swaps and tone flourishes.

| Wind | Lore | Color | Themes |
|------|------|-------|--------|
| **Hysh** | Light | White | Illumination, purity, exorcism, banishment of darkness |
| **Chamon** | Metal/Alchemy | Gold | Transmutation, forging, gilding, the philosopher's work |
| **Ghyran** | Life/Beasts (Jade) | Green | Growth, healing, seasons, the wildwood |
| **Ghur** | Beasts (Amber) | Brown | Primal savagery, shapeshifting, the call of the wild |
| **Azyr** | Heavens (Celestial) | Blue | Stars, prophecy, fate, lightning from above |
| **Ulgu** | Shadow (Grey) | Grey | Illusion, deception, misdirection, the unseen |
| **Shyish** | Death (Amethyst) | Purple | Endings, souls, the passage beyond, rest |
| **Aqshy** | Fire (Bright) | Red | Passion, destruction, courage, the cleansing flame |

### Warhammer vocabulary for the lexicon

- **Magister** — a licensed, college-trained wizard (vs hedge wizard)
- **Patriarch** — head of a College, equivalent to a dean or department head
- **Loremaster** — one who knows all spells of a lore
- **Hedge wizard** — unlicensed, self-taught magic user; the countryside hedge-witch
- **Channeling** — drawing raw magic from the Winds before shaping it into a spell
- **Miscast** — a spell that backfires; the magic escaping the wizard's control
- **Dhar** — raw, undiluted dark magic; forbidden and corrupting
- **Qhaysh** — High Magic; the harmonious weaving of all eight winds (elves only)
- **Articles of Imperial Magic** — the legal code governing wizard conduct
- **lingua praestantia** — the spellcasting language, derived from Eltharin
- **Thaumaturgical** — pertaining to the practical working of magic
- **Witch Hunter** — one who hunts unlicensed or corrupted magic users

### Warhammer tone flourishes (for the misfortune/triumph/aside pools)

Misfortunes:
- "The winds of magic turned against me."
- "Tzeentch's Curse, they call it. I call it Tuesday."
- "Reality stuttered, as it does when you prod it too hard."
- "The aether tasted of copper and regret."

Triumphs:
- "Even the Patriarch could not have done better. Well, maybe a little."
- "The Articles of Imperial Magic say nothing about this being forbidden. Yet."

## D&D — Schools of Magic

The eight D&D schools, useful for categorizing replacement vocabulary:

| School | Domain | Good for |
|--------|--------|----------|
| **Abjuration** | Protection, warding, banishment | firewall, antivirus, security |
| **Conjuration** | Summoning, teleportation, creation | containers, VMs, instantiation |
| **Divination** | Knowledge, scrying, foretelling | analytics, logging, monitoring |
| **Enchantment** | Mind-affecting, charm, compulsion | UX, persuasion, automation |
| **Evocation** | Raw energy, destruction, creation | compilation, rendering, GPU |
| **Illusion** | Deception, misdirection, disguise | CSS, theming, caching |
| **Necromancy** | Death, undeath, life force | garbage collection, deprecation |
| **Transmutation** | Transformation, change | refactoring, data conversion |

## Discworld (Terry Pratchett) — Unseen University

The tool's tone flourishes explicitly cite Pratchett-style bathos. Key UU
vocabulary:

- **Unseen University** — the premier school of wizardry (nobody actually teaches)
- **Archchancellor** — head of UU; traditionally short-tenured (assassination-based promotion)
- **Bursar** — the one person who understands the budget; kept sane with dried frog pills
- **Librarian** — an orangutan; refuses to be turned back ("OOK")
- **L-space** — all libraries connect through knowledge; time and space break down
- **Thaumaturgical** — the adjective for magic-engineering (the High Energy Magic Building)
- **Octiron** — a metal that is slightly impossible; Old Tom's bell is made of it
- **Sourcerer** — a wizard with raw magic in their blood; an apocalypse in a hat
- **Rite of AshkEnte** — summons Death; traditionally done at midnight with equipment

### Pratchett flavor lines (for the flourish pools)

Asides:
- "(the dried frog pills were not helping)"
- "(this was technically heresy, which is a kind of promotion at UU)"
- "(the Librarian would have been furious, had he understood any of it)"

Interjections:
- "By the Eightfold Path,"
- "In accordance with the Articles,"
- "As any student of the art knows,"

## Existing Lexicon Vocabulary Audit

Terms already in the lexicon that align with the above:

- "thinking engine" (computer) — evocation aesthetic
- "summoning altar" (server) — conjuration
- "great tome of records" (database) — divination
- "incantations" (code) — universal
- "grimoire" (codebase) — universal
- "curse" (bug) — necromancy/abjuration
- "curse-lift" (debug) — abjuration
- "dark omen" (error) — divination (negative)
- "transmutation circle" (compiler) — transmutation
- "scroll" (script) — universal
- "spell" (function) — universal
- "rune" (variable) — abjuration
- "oracle" (api) — divination
- "the great aether" (internet) — aqshy/celestial
- "sky realm" (cloud) — azyr
- "parchment" (file) — universal
- "secret word of passage" (password) — abjuration
- "scrying glass" (screen/monitor) — divination
- "familiar" (mouse) — conjuration
- "wizard" (developer) — universal
- "artificer" (engineer) — transmutation
- "bound spirit" (daemon) — necromancy

## Suggested New/Refined Entries

These fill gaps in the existing lexicon using the reference material above:

### Tech terms not yet covered
- "pipeline" → "ritual chain" (CI/CD, data pipelines)
- "container" → "sealed vessel" (Docker containers)
- "Docker" → "the artificer's vessel-lore"
- "Kubernetes" → "the great orchestration"
- "token" → "ward-stone" (auth tokens)
- "encryption" → "the cipher-ward" (already partially covered by "secret word")
- "authentication" → "the proving rite"
- "authorization" → "the granting of boons"
- "endpoint" → "nexus point"
- "webhook" → "aether-signal"
- "streaming" → "flow of thaumaturgical essence"
- "real-time" → "in the very moment" or "with nary a breath between"
- "dashboard" → "the scrying panel"
- "analytics" → "augury"
- "telemetry" → "divinatory readings"
- "metrics" → "omens measured"
- "scalable" → "that which swells to fill the need"
- "deployment" → "the unleashing" (already have deploy→unleash)
- "microservice" → "lesser enchantment" (already app→minor enchantment)
- "API key" → "key to the oracle's gate"

### Modern AI terms (already added — listed for the reference doc)
- "AI" → "arcane mind"
- "agentic" → "self-willed"
- "agent" → "emissary" (noun)
- "bot" → "automaton"
- "chatbot" → "talking head"
- "crawler" → "aether-prowler"
- "LLM" → "great mind"
- "prompt" → "invocation"
- "inference" → "divination"
- "fine-tuning" → "refining the bound mind" (potential phrase entry)
- "hallucination" → "phantom vision" (when AI fabricates)
- "context window" → "the circle of recollection"
- "embedding" → "soul-impression"
- "vector" → "directional rune"
- "RAG" → "consulting the great tomes mid-divination"

## World of Warcraft — Spell and Magic Vocabulary

WoW has the richest corpus of named magical abilities in gaming. Below is
organized reference material suitable for the lexicon.

### Mage — Arcane, Fire, Frost

**Arcane spells:** Frostbolt, Arcane Blast, Arcane Missiles, Arcane Barrage,
Arcane Explosion, Arcane Intellect, Arcane Surge, Counterspell, Blink,
Invisibility, Polymorph, Teleport, Portal, Conjure Refreshment, Time Warp,
Mirror Image, Spellsteal, Evocation, Touch of the Magi, Arcane Singularity,
Clearcasting, Prismatic Barrier.

**Fire spells:** Fireball, Fire Blast, Pyroblast, Meteor, Scorch, Combustion,
Cinderstorm, Pyroclasm, Living Bomb, Flamestrike, Hot Streak, Blazing Barrier,
Dragon's Breath, Phoenix Flames.

**Frost spells:** Frost Nova, Cone of Cold, Ice Lance, Blizzard, Flurry, Frozen
Orb, Comet Storm, Ice Block, Ice Barrier, Flurry, Brain Freeze, Fingers of
Frost, Glacial Spike, Ray of Frost, Ring of Frost, Summon Water Elemental.

**Useful mage vocabulary for the lexicon:**
- **Arcane Intellect** — a buff increasing intellect (study, learning)
- **Polymorph** — transform into a sheep (or turtle, pig, etc.)
- **Conjure Refreshment** — create food/drink from nothing
- **Time Warp** — temporal haste, a burst of speed
- **Spellsteal** — steal an enemy's enchantment
- **Evocation** — channel mana to restore power
- **Mirror Image** — create illusory duplicates
- **Counterspell** — silence/interrupt enemy casting
- **Portal/Teleport** — instantaneous travel
- **Blink** — short-range instant teleport

### Warlock — Affliction, Demonology, Destruction

**Core spells:** Shadow Bolt, Corruption, Drain Life, Drain Soul, Immolate,
Incinerate, Chaos Bolt, Fear, Banish, Curse of Weakness, Curse of Tongues,
Curse of Exhaustion, Howl of Terror, Shadowfury, Mortal Coil, Dark Pact,
Seed of Corruption, Unstable Affliction, Agony, Haunt.

**Summoning:** Summon Imp, Summon Voidwalker, Summon Sayaad (Succubus/Incubus),
Summon Felhunter, Summon Felguard, Summon Infernal, Summon Doomguard, Ritual
of Summoning, Ritual of Doom, Demonic Circle, Demonic Gateway.

**Resources:** Soul Shards, Healthstone, Soulstone, Soulwell, Soul Harvest,
Soulburn, Soul Link.

**Useful warlock vocabulary:**
- **Soulstone** — a gem storing the essence of a soul (for resurrection)
- **Healthstone** — a consumable healing item created from soul energy
- **Drain Life/Soul** — siphoning essence from a target
- **Corruption/Agony** — damage-over-time curse effects
- **Demonic Circle** — a placed ward allowing teleport back
- **Banish** — exile a demon or elemental from this plane
- **Subjugate Demon** — bend a demon to your will
- **Eye of Kilrogg** — a scouting summon (floating eye)
- **Ritual of Summoning** — group summoning of a distant player

### Priest — Holy, Discipline, Shadow

**Holy:** Heal, Flash Heal, Greater Heal, Prayer of Healing, Circle of Healing,
Holy Nova, Guardian Spirit, Divine Hymn, Sanctify, Serenity, Lightwell, Desperate
Prayer, Renew.

**Discipline:** Power Word: Shield, Penance, Atonement, Pain Suppression, Rapture,
Power Word: Barrier, Smite, Purge the Wicked, Evangelism.

**Shadow:** Mind Blast, Mind Flay, Shadow Word: Pain, Shadow Word: Death, Vampiric
Touch, Vampiric Embrace, Devouring Plague, Mind Sear, Shadowform, Dispersion,
Voidform, Void Bolt, Shadow Crash, Mindbender, Shadowfiend.

**Useful priest vocabulary:**
- **Power Word: Shield** — a protective ward spoken into being
- **Penance** — rapid channeled holy bolts (or shadow bolts)
- **Atonement** — healing through dealing damage
- **Mind Flay/Sear** — psychic assault on the mind
- **Vampiric Touch** — siphon life through magical contact
- **Dispersion** — dissolve into shadow to reduce damage
- **Divine Hymn** — a channeled holy song of healing
- **Shadowform** — assume a state of shadow-infused being

### Paladin — Holy, Protection, Retribution

**Paladin spells:** Holy Light, Flash of Light, Lay on Hands, Divine Shield,
Blessing of Kings, Blessing of Protection, Blessing of Freedom, Blessing of
Sacrifice, Hammer of Justice, Avenger's Shield, Judgment, Consecration, Holy
Shock, Light of Dawn, Templar's Verdict, Divine Storm, Shield of the Righteous,
Word of Glory, Aura Mastery, Avenging Wrath.

**Useful paladin vocabulary:**
- **Lay on Hands** — the ultimate healing touch (drains all mana)
- **Divine Shield** — total invulnerability ("bubbling")
- **Consecration** — sanctify the ground, burning the unworthy
- **Hammer of Justice** — stun with a thrown hammer of light
- **Avenging Wrath** — wings of light, a burst of power
- **Blessing of Freedom** — remove all movement impairing effects
- **Aura Mastery** — amplify the party's protective aura

### Death Knight — Blood, Frost, Unholy

**DK spells:** Death Strike, Death Coil, Death Grip, Obliterate, Frost Strike,
Scourge Strike, Plague Strike, Rune Strike, Howling Blast, Remorseless Winter,
Army of the Dead, Raise Dead, Death and Decay, Anti-Magic Shell, Icebound
Fortitude, Vampiric Blood, Blood Boil, Pestilence, Dark Transformation, Summon
Gargoyle, Unholy Blight, Chains of Ice, Empower Rune Weapon.

**Useful DK vocabulary:**
- **Death Grip** — yank an enemy toward you across the battlefield
- **Anti-Magic Shell** — absorb spell damage
- **Army of the Dead** — summon a swarm of ghouls
- **Remorseless Winter** — a freezing aura
- **Vampiric Blood** — siphon vitality through blood magic
- **Empower Rune Weapon** — refresh rune power in an instant
- **Death and Decay** — desecrate the ground, eroding all who stand

### Druid — Balance, Feral, Guardian, Restoration

**Druid spells:** Wrath, Moonfire, Sunfire, Starsurge, Starfall, Rejuvenation,
Regrowth, Wild Growth, Lifebloom, Healing Touch, Swiftmend, Innervate, Rebirth,
Entangling Roots, Cyclone, Barkskin, Ironfur, Frenzied Regeneration, Maul,
Swipe, Mangle, Rip, Ferocious Bite, Savage Roar, Moonkin Form, Cat Form, Bear
Form, Travel Form, Flight Form, Aquatic Form.

**Useful druid vocabulary:**
- **Innervate** — restore mana to self or ally
- **Rebirth** — resurrect a fallen ally mid-combat
- **Entangling Roots** — root a target in place with grasping vines
- **Cyclone** — banish a target into a cyclone (temporarily removed from battle)
- **Lifebloom** — a healing that blooms on expiry
- **Starsurge/Starfall** — celestial arcane bombardment
- **Shapeshift** — change form (bear, cat, moonkin, etc.)

### Evoker — Devastation, Preservation, Augmentation

**Evoker spells:** Living Flame, Fire Breath, Eternity Surge, Pyre, Deep Breath,
Emerald Blossom, Dream Breath, Spiritbloom, Time Dilation, Rewind, Temporal
Anomaly, Tip the Scales, Hover, Firestorm, Shattering Star, Engulf, Essence
Burst, Verdant Embrace, Breath of Eons, Upheaval, eruption, Quell.

**Useful evoker vocabulary:**
- **Essence** — the evoker's primary resource (like mana but more primal)
- **Tip the Scales** — empower the next spell to maximum instantly
- **Rewind** — reverse damage taken by the party
- **Living Flame** — fire that heals allies or burns enemies
- **Emerald Blossom** — burst of healing flora
- **Hover** — take flight momentarily while casting
- **Deep Breath** — soar across the battlefield, leaving devastation

### Shaman — Elemental, Enhancement, Restoration

**Shaman spells:** Lightning Bolt, Chain Lightning, Earth Shock, Flame Shock,
Frost Shock, Lava Burst, Lightning Surge, Earthquake, Elemental Blast, Stormstrike,
Lava Lash, Crash Lightning, Windfury Weapon, Flametongue Weapon, Healing Wave,
Chain Heal, Healing Surge, Riptide, Healing Rain, Spirit Link Totem, Spiritwalkers
Grace, Bloodlust/Heroism, Totems (Healing Stream, Mana Spring, Searing, Windfury,
Grounding, Capacitor, Tremor), Wind Shear, Purge, Ghost Wolf, Astral Recall,
Far Sight, Water Walking.

**Useful shaman vocabulary:**
- **Bloodlust/Heroism** — massive party-wide haste buff
- **Totem** — a placed magical ward providing an aura effect
- **Spirit Link Totem** — redistribute party health evenly
- **Chain Lightning/Heal** — effect that arcs between targets
- **Ghost Wolf** — shapeshift into a spectral wolf for speed
- **Water Walking** — walk across water surfaces
- **Wind Shear** — interrupt enemy casting
- **Healing Rain** — rain that heals allies standing in it

### WoW Item Suffixes (Classic Random Enchants)

Green-quality items in WoW Classic carry random suffixes. These are gold mines
for adjective replacements:

| Suffix | Stats | Wizardly use |
|--------|-------|-------------|
| of the Bear | Stamina, Strength | sturdy, resilient |
| of the Whale | Stamina, Spirit | enduring, calm |
| of the Eagle | Stamina, Intellect | wise, thoughtful |
| of the Boar | Stamina, Spirit | stubborn, vital |
| of the Gorilla | Strength, Intellect | brute cunning |
| of the Falcon | Agility, Intellect | keen-eyed, quick |
| of the Monkey | Agility, Stamina | agile, nimble |
| of the Tiger | Agility, Strength | fierce, predatory |
| of the Wolf | Agility, Spirit | pack-minded, alert |
| of the Owl | Intellect, Spirit | wise, observant |
| of Stamina | +Stamina | of endurance |
| of Intellect | +Intellect | of intellect |
| of Strength | +Strength | of might |
| of Agility | +Agility | of nimbleness |
| of Spirit | +Spirit | of spirit |
| of the Sorcerer | Stam, Int, Spi | well-rounded mage |
| of the Invoker | Int, Sta, Spd | power-caster |

### WoW Professions Vocabulary

| Profession | What it does | Wizardly term |
|-----------|-------------|--------------|
| Alchemy | Brew potions, transmute materials | the art of the crucible |
| Enchanting | Disenchant gear, enchant gear | the binding of wards |
| Inscription | Mill herbs, craft glyphs/scrolls | the art of sigils |
| Blacksmithing | Forge weapons and armor | the forge-master's craft |
| Leatherworking | Craft leather gear | the tanner's art |
| Tailoring | Craft cloth gear | the weaver's craft |
| Jewelcrafting | Cut gems, craft jewelry | the gem-cutter's lore |
| Engineering | Craft gadgets and explosives | the tinkerer's madness |
| Herbalism | Gather herbs | the gathering of simples |
| Mining | Mine ore and gems | the delving for earth-treasure |
| Skinning | Skin beasts | the flayer's trade |
| Fishing | Fish | the angler's patience |
| Cooking | Cook food buffs | the preparation of feasts |
| First Aid | Bandages | the field-healer's craft |

### WoW Tone Flourishes (for misfortune/triumph/aside pools)

**Misfortunes:**
- "My soulstone cracked. I have backups. I have backups."
- "The raid wiped. Forty gold in repairs. Worth it. Not worth it."
- "The Lich King himself would have winced. Actually, he did. I saw."
- "My mana bar was empty. I kept casting anyway. It did not work."
- "The dungeon finder gave me hope. The dungeon finder lied."
- "I stood in the fire. The raid leader was not surprised."

**Triumphs:**
- "Wipe on trash, one-shot the boss. Such is the way."
- "The loot dropped. It was not for my class. I needed it. I took it."
- "Achievement unlocked. Nobody saw it. That makes it better."
- "Full clear, no deaths. The healers demand tribute. Pay them."

**Asides:**
- "(the repair bill was catastrophic)"
- "(I should have brought more potions)"
- "(nobody reads the dungeon journal)"
- "(the auction house is its own kind of dark magic)"
- "(I once wiped a raid by accident. They do not know it was me.)"

### WoW-Inspired Lexicon Entries (suggested additions)

**Tech terms mapped through WoW vocabulary:**
- "sprint" → "ghost wolf" (shaman speed form)
- "dashboard" → already "scrying panel" — could also be "command totem"
- "shortcut" → "portal" (mage instant travel)
- "backup" → "soulstone" (warlock resurrection item)
- "rollback" → "rewind" (evoker temporal reversal)
- "cooldown" → "casting recovery" or "the mana to breathe"
- "spawn" → "summon" (generic conjuration)
- "loot" → "spoils of the encounter" (or "the boss drop")
- "AFK" → "away from the keyboard" — wizardly: "stepped through a portal"
- "patch notes" → "the chronicle of changes" or "new decree from the gods"
- "release" → "the unveiling" (already partially covered)
- "hotfix" → "a ward hastily drawn"
- "breaking change" → "a sundering of the old ways"
- "migration" → "the great exodus of the data"
- "staging" → "the proving grounds"
- "production" → "the live realm"
- "sandbox" → "the training dummies"
- "refactor" → "reshaping the incantations"
- "bug bounty" → "bounty on curses"

