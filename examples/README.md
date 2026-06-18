# Examples

This folder contains before-and-after transformations showcasing wizardify at different intensity levels.

## Tech Blog Post

Original: `tech-blog-original.md`

### Intensity 1 (Light Seasoning)
See: `tech-blog-i1.md`

Key changes:
- computer → thinking engine
- debug → untangle
- database → chronicle
- code review → scrutiny

### Intensity 2 (Enchanted)
See: `tech-blog-i2.md`

Adds:
- Multi-word replacements (machine learning → mechanical divination)
- Occasional interjections ("By my beard, the tests passed!")

### Intensity 3 (Full Wizard)
See: `tech-blog-i3.md`

Adds:
- Archaic grammar (thou/thee, -eth conjugation)
- Sentence inversions ("Never have I seen such elegance")
- Tone-aware flourishes (dark comedy after failures, pomp after wins)
- Advanced archaic: do-support removal, 'tis/twas

## How to Generate

```bash
# Light seasoning
../wizardify tech-blog-original.md -i 1 -o tech-blog-i1.md

# Enchanted
../wizardify tech-blog-original.md -i 2 -o tech-blog-i2.md

# Full wizard
../wizardify tech-blog-original.md -i 3 -o tech-blog-i3.md
```

## Adding Your Own Examples

1. Add a source file: `my-document-original.md`
2. Generate outputs at each intensity level
3. Update this README with a section describing the transformation
4. Commit all files together

---

## Example Transformations

### "Deploying to Production"

**Original (intensity 0):**
> Today we deployed the new API to production. The team tested thoroughly and we had zero bugs. Amazing work, everyone!

**Intensity 1:**
> Today we unleashed the new API upon the realm. The company scrutinized meticulously and we had zero hexed sprites. Remarkable endeavor, all!

**Intensity 2:**
> Hark! Today we unleashed the new API upon the realm. The company scrutinized meticulously and we had zero hexed sprites. By my beard, remarkable endeavor, all!

**Intensity 3:**
> Hark! Today did we unleash the new API upon the realm. The company did scrutinize most thoroughly and we had zero hexed sprites. Bards shall sing of this, barely.

---

### "Debugging a Complex Issue"

**Original:**
> I spent three hours debugging the memory leak. The issue was subtle and it took careful inspection to find it. I'm exhausted.

**Intensity 1:**
> I spent three hours unraveling the spectral drainage. The matter was mysterious and it took careful inspection to discover it. I'm depleted.

**Intensity 2:**
> Forsooth, I spent three hours unraveling the spectral drainage. The matter was mysterious and it took careful inspection to discover it. I'm depleted.

**Intensity 3:**
> In sooth, I spent three hours unraveling the spectral drainage. The matter was most mysterious and careful scrutiny did I undertake to discover it. I am grievously depleted. (Do not ask what it cost.)

---

### "Code Review Feedback"

**Original:**
> Please refactor this code. The function is too long and the variable names are confusing. After you fix this, we can merge.

**Intensity 1:**
> Please reconstruct this enchantment. The incantation is overly lengthy and the rune labels are bewildering. After thou fixest this, we can combine.

**Intensity 2:**
> I beseech thee, reconstruct this enchantment. The incantation is overly lengthy and the rune labels are bewildering. After thou fixest this, we can combine.

**Intensity 3:**
> I beseech thee most earnestly: reconstruct this enchantment! The incantation doth grow overlength and the rune labels perplex most grievously. Once thou hast fixed this matter, we shall merge into the chronicle. (And mark me well: this shall require thought.)
