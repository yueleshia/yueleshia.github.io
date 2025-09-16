{
  "title":   "Migrating Off My Custom Static Site Generator",
  "date":    "2025-09-22",
  "updated": "2025-09-22",
  "layout":  "post.shtml",
  "tags":    ["archive", "text"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

!{#run: BUILD=build tetra run file % >index.smd }

@TODO: Introduction

I have devoted a considerable portion of my life to language learning.
I've written an undergraduate thesis in Mandarin, and I've gained Japanese fluency mostly through media.
And maybe two more languages will join the fluency tier.
I've always had a passion for writing, and one day I thought, why not challenge myself to write about my interests in multiple languages.

# The Journey to AsciiDoctor

## Digitising My Notes

I had always enjoyed long-form creative and technical writing.
But this journey to my own tech stacks begins a few months after learning Vim, when I began looking into more effective note taking.

GUI programs like EverNote and Notion always felt like too much machinery and abstraction for something as simple as notes.
The beauty of simple text files is how easy it is craft a short script to do something that your note take app never imagined.
Like navigating to the next link in Vim, finding backlinks, link checking with [lychee](https://github.com/lycheeverse/lychee), etc.

At some point I discovered VimWiki, which reminded me a lot of my Wikipedia editing days.
I discovered `gf` (goto file) where you could navigate within Vim to locally linked files for a web-browser-like experience.
Blogging is as much for me (as a reference) as it is for others, and if I had to have this feature if I were to pick up blogging again.

## Markdown

There's so much to love and hate about it.
I began to put all writing and notes into Markdown.
It contains all the primitives for most types of writing in extremely concise syntax, perfect for distraction free writing.[^wordstar]
It is a durable spec because of its simplicity, and it is amenable to Linux CLI tools because of how it handles single/double newlines (only two newlines are a paragraph break).

[^wordstar]: Distraction free tools are important for writing because they let you remain in the flow state. 
[Neal Stephen](https://en.wikipedia.org/wiki/Neal_Stephenson) writes an essay, ["In the Beginning... Was the Command Line"](https://web.stanford.edu/class/cs81n/command.txt), about how Emacs allows professional writer to focus on writing and make "everything else [like formatting and printing] vanish".
[Robert J. Sawyer](https://en.wikipedia.org/wiki/Robert_J._Sawyer) writes [about choosing Wordstart](https://www.sfwriter.com/wordstar.htm) mentioning its distraction-free UI.
Amusingly, this write-up echoes the exact sentiments of why people choose TUI text editors like Vim/Emacs/helix and extol the virtues of Libre software.
[George R. R. Martin](https://en.wikipedia.org/wiki/George_R._R._Martin) talks about unwanted features in Modern word processors, like spell check on [Conan O’Brien's show](https://www.youtube.com/watch?v=X5REM-3nWHg).

However, Markdown is not without its warts.
Its spec, while simple, is so underspecified that it spawned the [CommonMark movement](https://commonmark.org/).
Its minimal feature set, while a strength, is also its weakness because it two or three features (specifically: tables, anchor links to headers, and footnotes).
Many of these features are available as extensions to CommonMark, but the world could have been so much better if they were part of the original Markdown standard.

No footnotes were the deal breaker for me.
Footnotes are critically important for education and dissemination of information.
They provide the option a reader (and the writer) to revisit the original context and their own critical evaluation of arguments presented.
They are invaluable as tool for adding tangential information without sacrificing reading clarity.
And most of all, unlike other features, adding them through simple scripts is infeasible due to footnote backlinks and auto-numbering.

## Settling for AsciiDoctor

Over the years, I tried a lot of technologies.

I tried [Jyupter Notebook](https://jupyter.org/try) for a data science class, but abandoned it when I couldn't use anything but Markdown for text.
I tried [Groff](https://www.man7.org/linux/man-pages/man7/groff_man.7.html) as I was learning about Linux, but it lacks features.
I tried [Pollen](https://docs.racket-lang.org/pollen/), but abandoned it once I understood I was being hoodwinked into learning an entire Lisp just for writing.
I learned [LaTeX](https://en.wikibooks.org/wiki/LaTeX) and [RMarkdown](https://bookdown.org/yihui/rmarkdown/) for writing my thesis and later for some contracting work, but the machinery is a little too brittle for my tastes.

At some point I learned about the method-taking framework Zettelkasten.
For me the key takeaway of Zettelkasten for tooling is the importance of backlinks.
[VimWiki](https://github.com/vimwiki/vimwiki) and [Corpus](https://github.com/wincent/corpus) provided just that, but were limited to Vim.
[Obsidian](https://obsidian.md/) and [LogSeq](https://logseq.com/) had the opposite problem being limited to the GUI.
But I want notes that I can view both on my website, in Vim, on phone, and on a GUI with navigation in each medium.

Eventually, I settled on [AsciiDoctor](https://asciidoctor.org/).
It has all the features I want (and many I do not), and ultimately feels very close in spirit to Markdown.

Sometime after, the CommonMark movement started.
That's pretty cool.

# The Dream Multilingual Workflow

Perhaps unsurprisingly, a single author acting as the source of manuscript of two translation is not a use case supported anywhere.
I imagine having a single text file where you interleave paragraphs with multiple languages.

```
= Example Source Document

{{ en }}

I imagined writing documents that looked like this: different languages interleaved in the source document.
You would have a single paragraph in English.

Or a list of paragraphs in English.
Followed by their translated versions.
These would be marked with labels for which language that could be used to render target-language documents.

{{ jp }}

複数の言語が交じあって本書同様な文書を書くことを想像した。
初めに英語の段落があるでしょう。

または、複数の英語段落。
続いては多言語の文章。
これらは、各断面に言語の印を置くことで各翻訳先きの文書が創作可能。

{{ * }}

I would also support code blocks that would be included by every language render.

[source,zig]
----
pub fn main() void {
    // some random code
}
----
```

In fact, the implementation of this was my very first [Rust project](https://github.com/yueleshia/polyglot).

# Building My Own Markdown (Tetra)

Perhaps it all started when I learned that you could dataviews in Obsidian.
Or perhaps it was because I wanted Jyupter Notebook but for AsciiDoctor.
But at some point, I thought to myself: I write in a lot of programming languages that are good for different things... why can't I just embed them into AsciiDoctor?

And so I began [Tetra](https://github.com/Yueleshia/tetra).
At first it looked something along the lines of this:

```
{| sh tr '[[:lower:]]' '[[:upper:]]' |}
{% echo hello %}
word
```

You would inline commands `echo hello` and block commands `tr` that take in a block of text as STDIN.
This snippet would resolve to `HELLO WORLD`.
(This is very similar to what it is today.)

I also created a simple lisp-like language for it (which is no longer part of `tetra`), which was my first time writing a compiler for a programming language.
My primary use case was to implement citations (e.g. `{% cite DeFrancis 3 %}`) via LaTeX's BibTex, and a `{% references %}` command that make the bibliography.
And it worked (via [pandoc](https://github.com/jgm/pandoc)).

## Experimenting Memory Architecture

An interesting property of how I wrote Tetra is that, given an input string, you can calculate the memory upper bound without doing full parsing.
It dawned on me that I could reuse the function I use for parsing to generate calculate an exact size with macros.[^macro-parsing]
I would then assert that final output could not be greater than my precalculated capacity.

[^macro-parsing]: I'm definitely not proud of the code, but [here](https://github.com/yueleshia/chordscript/blob/main/chordscript/src/deserialise/keyspace_preview.rs).
I definitely would do things like this anymore.
I would show in the [Tetra codebase](https://github.com/yueleshia/tetra/commits/rust/), but I'm lazy to find it.
It exists somewhere in the commit history on the `rust` branch.

I eventually realised that was the very opposite of readable, but I still really liked the concept of precalculating the capacity based on the input string.
At some point, perhaps from watching [Handmade Hero](https://hero.handmade.network/), I learned about grouped memory management.
Rather than allocating and freeing memory 1-by-1 for every variable decleration (ie. what the Rust borrow checker is automating), one would less granularly group several allocations together.
And so began the third rewrite of Tetra.
I also wrote unsafe Rust for the first time to implement my Arena.

## Unlimited Rewrites

Although I really wrestled with lifetimes, the Arena rewrite was mildly successful.
But then I thought, why not have the language splitting feature as well?
So it began again.

Well, suffice to say, I have written tetra many times since.
In my most recent rewrites, I moved to Zig and tried out table-driven parsing (see this post where I talk about it).
But I think I'm mostly satisfied with its architecture now, and it is what powers this blog.
In Zig 16.0 (or sometime [soon](https://ziglang.org/devlog/2025/#2025-06-30)), Zig's colourless async interface will land (in what feels very much like an IO monad in functional programming), which let me to finally parallelise Tetra's execution.

One thing that I did lose through all these rewrites was the lisp-like scrpiting language.
It ended slowing down the design work too much.
Maybe I will return to it in the form of [EYG](https://github.com/yueleshia/eyg), or maybe I will settle for Lua.

# Static Site Generators

I built my markdown language to fill in any gaps in any Markup language I happen to choose.
And it can split languages if I so wish.
Now, let's make a website out of it.


## Shellscript

For this use case, at its core, a static site generator compiles markup into HTML, rebases links to your website's domain, and provides you a way to reuse HTML (e.g. have the same navbar across all pages).
With a healthy allergy to bloat, I thought to myself, I can just make a few liner script that automates it.
In fact, I rewrote the implementation more than once, and tried a few single-file bash scripts that others had made.

Through all the rewrites, I think the part that was most difficult was always ensuring links worked locally if I previewed them directly with `file://`, used `gf` in Vim, or used a webserver with `http://localhost/`.

## Solving "Not Invented Here" with Hugo

> If it is not essential to the business, then delegate it to a third party.

Through all the rewrite churn, I thought to myself: I am not really adding anything to site generation, it would just make more sense to focus on iterating on `tetra` and use an existing static site generator.
I decided on Hugo because Go is statically compiled (so no extra dependencies), I can `apt install` it on GitHub runners, and Hugo supports AsciiDoctor. 

## Zine

I had always heard of Zine, but it only supports Markdown.
But once I learned it supported syntax highlighting via tree sitter, I knew I had to switch.
I always desiped having to install random python packages if I wanted to use pygments or ruby if I wanted to use rouge.
These are dependencies I much rather not have in my local computer, and the languages I use are sometimes rather esoteric (e.g. [nickel](https://www.github.com/tweag/nickel)).
I haven't looked back since.

# What Is Next?

Ironically, I am back to Markdown (CommonMark), but enhanced with Tetra.
[Typist](https://github.com/typst/typst) is the new kid on the block, and it is LaTeX replacement we have all be waiting for.
I write my CV in it, but it compiles too slowly for me for blog usage (few seconds on cold start).

I often wonder if Typst's syntax is the correct design for my use case.
It essentially has two modes, text mode or script mode, and you can nest them arbitrarily.
But I think I would rather write a Typst implementation rather and keep Tetra's design flat.
It does exactly what I need and the next design problem to tackle is parallelizing execution.

I never thought my small blogging language would turn into a multi-rewrite language.
Right now Tetra can spit out its AST, which I can then use to syntax check all my code snippets.
Unexpectedly (and some podcasts later), Tetra is growing into a language that can solve formal specification issues.[^formal-specification]
I certainly learned a lot from trying to solve a seemingly begnin problem.

[^formal-specification]: I am not certain that "format specification" is the correct term for it, but in high saftey environments, you have a requirement that needs to be linked to documentation and to specific lines of source code.
Testing code sinppets in documentation (as you can do in Rust) is the of the same praxis, just with much less rigour.

Keep scratching your developer itch, you never know where it will take you.

