{
  "title":   "International Text: How hard could it be?",
  "date":    "2025-09-21",
  "updated": "2025-09-21",
  "layout":  "post.shtml",
  "tags":    ["archive", "text"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

This is post I wish I had when I first started programming.
By no means do I know most of the stack, but I'll try to update this as I learn more.
You should also checkout this [write-up](https://www.newroadoldway.com/text0.html) by the Jimmy Lefevre, creator of the text shaping library [kb](https://github.com/JimmyLefevre/kb).


# The Seven Stages of Texting Grief

1. Graphics seems really complicated, and you've been printing text since you first learned programming, how hard can it really be?
1. Then you play with Go/Rust/Java with a Unicode character type or have to parse non-English text, and now you have to learn UTF8/UTF16.
1. Then you decide to only support ASCII or ignore Unicode behaviour: it's a breeze.
1. Then you learn graphics programming, and actually only the initialization and debugging is hard, but you'd stick to text if you could.
1. Then you start making a text interface with a language and now you have to care multilingual graphical text concerns (e.g. does I want a row of Japanese text and a row of English text to align), and you start having to dig into Unicode or outsource it to a library.
Making a game might be simpler than this.
1. Then you start making a Unicode library, text shaping library, or font.
1. Then you find the Microsoft spec for text shaping documentation is wrong nearly everywhere.

# Input Layer

Although this is mostly out of scope, I've included here for completeness.

In modern operating systems, there are three layers to the current text input:

* **Keyboard firmware**: Fancier keyboards have on-board software that will let you remap buttons to other keys.
For example, see the flashable [QMK](https://qmk.fm/).
* **Keyboard layouts/maps**: This is an operating-system-layer concept.
Typically countries have country-specific physical layouts designed for the national language, and keymaps ([XKB](https://wiki.archlinux.org/title/X_keyboard_extension) on Linux) are the abstraction for that.

* **Input method Editor** (IME): These UI around facilitating text input, especially for more complicated text.

Note that this only makes sense in the GUI. The flow is as follows:[^keyboard-flow] [^ime-flow] [^sdl-ime-flow]

1. The keyboard controller sends events[^key-event-bad-term] -> kernel
1. kernel -> evdev -> libinput -> Xorg/Wayland (remaps based on the XKB config files)
1. Xorg sends key events and keysym -> your program
1. Your program chooses:[^raylib-ime-pr]
    1. program send key events -> IME sends `text` -> input handling code
    1. program send key events -> input handling code

[^keyboard-flow]: [Wikipedia article on Linux's evdev](https://en.wikipedia.org/wiki/Evdev) showing the flow from keyboard -> window compositor.
[^sdl-ime]: [How SDL handles IME input.](https://wiki.libsdl.org/SDL2/Tutorials-TextInput)
[^ime-flow]: [Fcitx's contributor intro page](https://fcitx-im.org/wiki/Basic_concept) explaining IME concepts.
[^key-event-bad-term]: @TODO: I am not sure what the correct term for keyboard output.
[^raylib-ime-pr]: [A pull request for raylib IME support](https://github.com/raysan5/raylib/issues/1945) where they talk about how to include it and support it.

When executing a program (via a shell or via fork in a program), text is sent to your program as stream of bytes via STDIN.
In GUIs, the `window` (Win32/Xorg/Wayland) sends key events which you typically.

# Model Layer (Code Points)

We are not yet at the layer where text is stored (encoded) in memory or on disk; this will be addressed in the [](#data-layer).
At this layer, the Unicode consortium solves the problem of how do we represent all the languages in the world.

A `code point` is the index of an array where all the values are semantically unique: one large column.

* ASCII code point table is `[128]u7`
* Unicode code point table is `[1114112]u21` (with some unmapped values as it is an evolving standard)
* Others: Here is a survey of [character encodings used in the web](https://w3techs.com/technologies/overview/character_encoding), and one further broken down [by top-level domain](https://w3techs.com/technologies/segmentation/tld-jp-/character_encoding)

This not only includes characters like `0`, `a`, `B` that are encoded but also characters like `\n`, grave accent, zero-width joiner, smiling emoji, etc.
Unicode is an evolving standard, and to this today new characters are being added, and of course there are errors.

Some languages have a specific type that is semantically maps on to a code point like Go (`rune`), Rust (`char`) that are represented as u32.

## LOCALE

A related but separate topic is locale.
I would describe this as C's standard library API[^C-std-as-system-config]: glibc (and I assume msvc) supports it whereas musl does not.[^glibc-musl-locale] 
I believe locale percolates across programs because the C std lib is a transitive dependency of nearly every program.
But it is unrelated to text encoding/decoding.

Locale means things like date formats, currency symbols, decimal place characters, etc.
On Linux, you may be familiar with the environment variables [LANG, LC_CTYPE, LC_ALL, etc.](https://wiki.archlinux.org/title/Locale)

[^C-std-as-system-config]: All mainstream Operating Systems (OS) are built with the assumption of a specific C standard, and C std lib becomes part of the OS API.
[^glibc-musl-locale]: https://docs.voidlinux.org/config/locales.html

[]($section.id('data-layer'))
# Physical Layer (Strings)

Now we are at the in-memory or on-drive representation of text.
One could imagine a world where we are dealing with a stream of structs instead, but text data is de-facto handled as a stream of bytes.
And bytes (u8) form the basic unit of access for CPUs.

ASCII code points are u7, so you can do the obvious direct mapping from code point to u8.
For example, "Hello" is `['H', 'e', 'l', 'l', 'o']` which is `[72, 101, 108, 108, 111]` in a strongly-typed language.

But as Unicode is u21, which comes the complications of spanning across multiple bytes.
There are several ways that a code point can be encoded to a byte array, none of which are the obvious u24.
This is because Unicode different encodings design around the following constraints as it evolves to include more languages and characters:

* Backwards compatibility with itself and with ASCII.
* Size. UTF8 is the best size for most non-East-Asian texts, UTF16 is the best for East-Asian texts.
* Endianess (Byte Order Mark)
* Error correction (if a hard drive is corrected for example)

This [mail](archive-comparing-utf-encodings) is a nice comparison of the different encodings.
In the modern day, we only really have to interface with ASCII/UTF8 (probably because Linux won) and UTF16 (in Windows file paths and in JavaScript but not HTML).

## Code Points Encoded

UTF8 works by bit shifting the code point into a variable length container.
The first byte is either ASCII or tells you the length of 

1. U+0000 to U+007F are encoded directly as one byte (equivalent to ASCII)
1. U+0080 to U+07FF are encoded as a two-byte sequence
1. U+0800 to U+D7FF are encoded as a three-byte sequences
1. U+D800 to U+DFFF are invalid code points reserved for 4-byte UTF-16
1. U+E000 to U+FFFF are encoded as three-byte sequences same as above

UTF16 works by similarly but only for 

1. U+0000 to U+D7FF are encoded as two-byte 0x0000 to 0xD7FF
1. U+D800 to U+DFFF are invalid code points reserved for 4-byte UTF-16
1. U+E000 to U+FFFF are encoded as two-byte 0xE000 to 0xFFFFh
1. U+10000 to U+10FFFF use 4-byte UTF-16 by:
    1. Code point subtract 0x10000 (leaving you a u20)
    1. An bit shift the result into 110110xxxxxxxxxx 110111xxxxxxxxxx<sub>u16-bin</sub>.

Endianess within a byte however is something we never have to handle so individual bytes are displayed in the customary language order with largest bit on the left.
And, indeed, UTF8 is free of endianess issues.

If you wish to check yourself, run the following in python:

```python
'囧'.encode('utf-8')
'囧'.encode('utf-16-le')
'囧'.encode('utf-16-be')
```

The following are in byte the order as you would write them out as a string in source code.
However, the order of bits within a byte is constant (128-bit first) for ease of comparison and because generally it is only when dealing with networking that input is not already bit swapped for you.

Encoding                   | byte 0   | byte 1   | byte 2   | byte 3
---                        | -        | -        | -        | -
'z'                        | <code><span style="color: blue">01111010</span> (7A)</code> ||| [U+007A]
ASCII                      | <code><span style="color: blue">01111010</span> (7A)</code>
UTF8                       | <code><span style="color: blue">01111010</span> (7A)</code>
UTF16 LE                   | <code><span style="color: blue">01111010</span> (7A)</code> | <code>00000000 (00)</code>
UTF16 BE                   | <code>00000000 (00)</code> | <code><span style="color: blue">01111010</span> (7A)</code> |
UTF32 BE                   | <code>00000000 (00)</code> | <code>00000000 (00)</code> | <code>00000000 (00)</code> | <code><span style="color: blue">01111010</span> (7A)</code> |
&nbsp;|
'α' (greek alpha)          | <code>00000<span style="color: green">011</span> (03)</code> | <code><span style="color: blue">10110001</span> (B1)</code> || [U+03B1]
ASCII                      | N/A |||
UTF8                       | <code>110<span style="color: green">011</span><span style="color: blue">10</span> (CE)</code> | <code>10<span style="color: blue">110001</span> (B1)</code> ||
UTF16 LE                   | <code><span style="color: blue">10110001</span> (B1)</code> | <code>00000<span style="color: green">011</span> (03)</code> ||
UTF16 BE                   | <code>00000<span style="color: green">011</span> (03)</code> | <code><span style="color: blue">10110001</span> (B1)</code> ||
&nbsp;|
'囧' (zh jiǒng)            | <code><span style="color: green">01010110</span> (56)</code> | <code><span style="color: blue">11100111</span> (E7)</code> || [U+56E7]
ASCII                      | N/A |||
UTF8                       | <code>1110<span style="color: green">0101</span> (E5)</code> | <code>10<span style="color: green">0110</span><span style="color: blue">11</span> (9B)</code> | <code>10<span style="color: blue">100111</span> (E7)</code> |
UTF16 LE                   | <code><span style="color: blue">11100111</span> (E7)</code> | <code><span style="color: green">01010110</span> (56)</code> ||
UTF16 BE                   | <code><span style="color: green">01010110</span> (56)</code> | <code><span style="color: blue">11100111</span> (E7)</code> ||
&nbsp;|
'🌈' (rainbow)             | <code>000<span style="color: red">00001</span> (01)</code> | <code><span style="color: green">11110011</span> (F3)</code> | <code><span style="color: blue">00001000</span> (03)</code> | [U+1F308]
ASCII                      | N/A ||| 
UTF8                       | <code>11110<span style="color: red">000</span> (f0)</code> | <code>10<span style="color: red">01</span><span style="color: green">1111</span></code> | <code>10<span style="color: green">0011</span><span style="color: blue">00</span></code> | <code>10<span style="color:  blue">001000</span></code>
UTF16 LE                   | <code><span style="color: red">001</span><span style="color: green">11100</span> (3C)</code> | <code>110110<span style="color: red">00</span> (D8)</code> | <code><span style="color: blue">00001000</span> (08) </code> | <code>110111<span style="color: green">11</span> (DF)</code>
UTF16 BE                   | <code>110110<span style="color: red">00</span> (D8)</code> | <code><span style="color: red">001</span><span style="color: green">11100</span> (3C)</code> | <code>110111<span style="color: green">11</span> (DF)</code> | <code><span style="color: blue">00001000</span> (08) </code>

It was useful to know this when working with WASM (UTF8 strings in source) and JavaScript (UTF16).

## Grapheme Clusters

@TODO

## Working with ASCII

Just some nice things about the design of ASCII.

* `'a' & 0x110111` uppercases
* `'A' | 0x001000` lowercases
* Because UTF8 is compatible with ASCII, you can quick test for Unicode with SIMD by `& 0x1000000`.
* For UTF8

And if it is ASCII, then you know the length of the string.
If you can assume English, then you can this massively simplifies your code.
You


# Visual Layer (Text Shaping)

## Fonts
### Bitmap Fonts

Perhaps the most obvious solution is to just join rectangles images for each glyph.
This is exactly what bitmap fonts are.
However, working with bitmaps leads to blurring (of the kind that isn't good for text)when scaling for different resolutions.
One solution could be [mipmapping](https://en.wikipedia.org/wiki/Mipmap), but Adobe pioneered vector fonts. 

### First Vector Fonts (Adobe postscript fonts)

These fonts represent the outline of characters with bézier or cubic curves, solving the scaling for font size issue.
Additionally, these fonts contained a series of postscript procedures called _hints_ to further refine characters, e.g. for different display resolutions.[^type1-hints]

[^type-hints]: Fishler, Ori. "Type1 Fonts". University of Waterloo. [https://cs.uwaterloo.ca/~dberry/COURSES/electronic.pub/fishler/type1.htm#hints](https://cs.uwaterloo.ca/~dberry/COURSES/electronic.pub/fishler/type1.htm#hints)

### TrueType Fonts (ttf)

However, vector fonts were not enough, and apple created TrueType to support:

* Font families (e.g. having regular/italics/bold in one file)
* Kerning (spacing glyphs non-uniformly, e.g. 'i' is a very thin letter compared to 'w')
* Avoiding adobe's licensing fees[^truetype-history]
* Font hinting

The kern table is represented as a n-squared table encoding the spacing between every pair of glyph within a font.
Thankfully, the reality of languages today do not require kern table combinatorial explosion.
Fonts are language-specific, and character inventory of a language is typically quite small.
CJKV languages are essentially monospaced and do not require kerning.[^cjkv-kerning]

[^truetype-history]: Penny, Laurence. "A History of TrueType". TrueType typography. https://www.truetype-typography.com/tthist.htm

[^cjkv-kerning]: It is not actually strictly true that all characters are the same size.
Japanese has three Japanese-specific writing systems, Han characters, hiragana, and katakana.
In analog writing, which begot and form the basis of the digital, hiragana and katakana are supposed to be smaller than Han characters.
One could imagine kerning this, but this is typically handling by down-sizing the glyphs relative to their Han-character counterparts.

### OpenType Fonts (otf)

However, this was not powerful enough to represent all writing systems.
Enter OpenType.

Arabic letters have four forms depending on the location within a word.
English cursive scripts bear this property as well.
Indic scripts have syllables as the base and characters combine in non-linear ways.

Thus you can think of OpenType as vector outlines (and hints) + kern tables + glyph substitution tables + glyph positioning tables.

## Shaping Engines

Unfortunately, OpenType is still not powerful enough.
Enter shaping engines (DirectWrite, [HarfBuzz](https://github.com/harfbuzz/harfbuzz), [kb](https://github.com/JimmyLefevre/kb)).
Microsoft, the publishers of the text shaping spec, document hand-coded

Unfortunately, for because the Unicode spec has errors, the Microsoft font spec has errors, the OpenType model is not powerful enough, language-specific details.[^lefevre]

[^lefevre]: Lefevre, Jimmy and Ścigaz, Łukasz. "The Ridiculous World of Text". YouTube, Wookash Podcast. Aug 16, 2025. [https://www.youtube.com/watch?v=hGvInFf839U](https://www.youtube.com/watch?v=hGvInFf839U)

## Feature: Font Fallback

@TODO

# Text Segmentation

* Split into `runs` (Uniform text and direction)

## @TODO: How is Han unification handled?

Han characters, like the Latin alphabet, became the writing script of a few languages.
Each language has its own story of changing how to write characters.[^nei] [^note-simplication]
The Wikipedia article on [Han Unification](https://en.wikipedia.org/wiki/Han_unification) gives a table.

[^nei]: An imperfect analogy for 內 vs 内 is if the Cyrillic and Latin alphabet were nearly identical and `b` vs `б` were region variations of the same letter.
[^note-simplication]: Typically, these language variants arose out of somewhat independent simplification processes.

I am unsure how a Japanese/Mandarin mixed body is suppose to be handled.
This exact characters might not be correct, but I recall once when typing out `真` in Discord (browser-based chat application), adding and deleting `と` (jp) right after was real-time swapping between the Mandarin and Japanese variants of the character.

# Conclusion: The Algorithm

As someone who has examined the source code of text shapers, I believe a full text rendering algorithm looks like the following:

1. Extract ANSI escape codes if for a TUI
1. Read in []u8
1. Decode Unicode into code points []21
1. Segment text based on Unicode's Character Database (USD) CSVs: [grapheme clustering, word breaking](https://www.unicode.org/reports/tr29/), [script breaking](https://www.unicode.org/reports/tr24/), [line breaking](https://www.unicode.org/reports/tr14/))
1. Segment text into runs
1. Customize
1. Shape runs with a given font and character class (obtained from the UCD).
1. Rasterise and render with BiDi in mind.

Because shaping is quite intensive, typically characters are stored in a least-recently used cache.
I'm curious how a LRU cache looks like for something like Arabic, since your key cannot be a codepoint.

If you can make the assumption about what languages you are supporting, then maybe you can find string length by taking the array length; if not then you will want need unicode decoding.
If you want to deal with text sizing for TUIs, then you need USD information for grapheme clustering and knowledge on the terminal emulator support for grapheme clustering.
See the creator of ghostty, Mitchel Hashimoto's [post](https://mitchellh.com/writing/grapheme-clusters-in-terminals) on the subject.
If you can assume English (which you often shouldn't for GUIs), then you can happily parse ASCII.
Otherwise you will need a unicode character data, text segmentation, and text shaping.

# External Links

* James Breen on "[Kanji and the Computer: A Brief History of Japanese Character Set Standards](https://sino-platonic.org/complete/spp360_kanji_computers_japanese_character_set.pdf)".
Japan's case is fairly interesting, because it was the largest market during the period before ASCII and Unicode had been codified.[^second-largest-market]
In particular, the 常用漢字 are the set of characters that learners are taught today.

* https://who-t.blogspot.com/2016/05/the-difference-between-uinput-and-evdev.html
* https://github.com/wez/evremap
* Wookash Podcast, [The Ridiculous World of Text | Jimmy Lefevre](https://www.youtube.com/watch?v=hGvInFf839U) on reimplementing HarfBuzz from scratch.
* If you'd like to learn about https://ebixio.com/online_docs/UnicodeDemystified.pdf
* [A Brief look at Text Rendering](https://www.youtube.com/watch?v=qcMuyHzhvpI) - A visual introduction to font encoding
* https://unicode.org/mail-arch/unicode-ml/y2007-m01/0057.html
* https://github.com/JimmyLefevre/kb
* https://github.com/n8willis/opentype-shaping-documents

[^second-largest-market]: https://en.wikipedia.org/wiki/Japanese_language_and_computers
