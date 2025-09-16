{
  "title":   "Table-Driven Lexing",
  "date":    "2025-06-29",
  "updated": "2025-09-24",
  "layout":  "post.shtml",
  "tags":    ["archive", "text"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

I always heard that table-driven parsing was optimal, and in theory I knew it was possible, but I never really understood how.
Well, past self, here is how it is done.

# Introduction to Compilation

If your input is the highest level-representation, 'compiling' in computer science is the process of taking some input and progressively lowering it into other representations.[^lowering]
For example, the Rust compiler transforms:

[^lowering]: The term 'lowering' probably arose from how High-Level Languages evolved out of Lower-Level Languages and required a more formalised process of parsing.

| Rust[^rust-lexing] [^zig-lexing1] [^zig-lexing2] | Zig
|---|---|
| source code (text)                           | source code (text)
| tokens via the __lexer__                     | tokens by the __tokenizer__
| AST via the __parser__                       | AST via the __parser__
| High-level Intermediate Representation (HIR) | Zig IR by __AstGen__
| Typed HIR                                    | Analyzied IR by __Sema__
| Mid-level A IR                               | Machine IR by __CodeGen__
| https://en.wikipedia.org/wiki/LLVM[LLVM] IR  | LLVM IR (when the LLVM backend is used)
| Machine code.                                | Machine Code

[^rust-lexing]: Rust Compiler Development Guide: Source Code Representation. <https://rustc-dev-guide.rust-lang.org/part-3-intro.html>
[^zig-lexing1]: Lugg, Matthew. link:https://youtu.be/KOZcJwGdQok?t=822[Data-Orientated Design Revisited: Type Safety in the Zig Compiler - Matthew Lugg]. 13:42. Software You Can Love, Milan 2024. Archived at YouTube, Zig SHOWTIME on 2024-09-05.
[^zig-lexing2]: Hashimoto, Mitchell. Zig. Mitchell Hashimoto, 2025. <https://mitchellh.com/zig>


Each step is essentially an Array/Tree of structs, and compilation is iteratively taking one representation and transform it into the next until you have your output.
More specifically, each step in this process is categorizing your input with types (structs), but the structs you use are often ambiguous until further down the pipeline, so each step is just categorizing it better and better.
And at each step, different operations become more feasible, e.g. type inference, optimizations, etc.

== Approaches to Lexing

The specific names for the lower processes, like lexing or parsing, may be called different things or may not even be a step in every compiler.
But, __lexing__ in compilers does typically always refer to grouping source-code characters together into more a useful token struct.
__Parsing__ often refers to the taking those tokens and transforming them into an AST, but sometimes it refers to __lexing__ and transforming into an AST combined.
There is nothing magical about tackling the problem in this way; parsing and lexing just happen to tractable enough abstraction.

## [Lex](https://en.wikipedia.org/wiki/Lex_(software))/[Flex](https://en.wikipedia.org/wiki/Flex_(lexical_analyser_generator))/[Yacc](https://en.wikipedia.org/wiki/Yacc)/[Pest](https://pest.rs/)/etc.

After writing a document for your project in a custom domain-specific language (DSL), these programs will do the lexing for you.{wj}footnote:[
Yacc and Pest are do more than just lexing.
Yacc uses the infamous Backus-Naur Form of languages that you often on many languages.
For example, link:https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_10_02[POSIX sh (search 'Shell Grammar Rules')] and link:https://ziglang.org/documentation/master/#Grammar[Zig (search 'Grammar')].
]
However, using these tools often is more pain than it is worth, especially for the older tools.
See https://www.nothings.org/computer/lexing.html[this article] by Sean Barrett on arguing against lex/flex.

Moreover, writing a lexer only takes a few hours, roughly how long it will take to understand how to integrate an external tool/library into your build.{wj}footnote:[
The parser into non-typed AST usually takes me a few days, but this is usually because this where I am also making decisions on language design.
]
Is it really worth adding a dependency just to save a few hours?
And even if you deem it so, is it really worth learning a new DSL just, that you probably will not touch again and forget, just to save those few hours?

## Handmade Switch-Based Lexing

A lexer (and indeed the parser) essentially are just a finite state machine.
You have some enum for you state, you loop over the input text, you switch on the state, then you switch on the text, and then update the state and maybe output.

Generally, the advantage of handmade lexers is that you can tailor your error messages much better.
Compiler error messages are the UI of your compiler, so it is fairly import to get them right.
There's a reason why Rust values error messages and was inspired by Elm's error messages.{wj}footnote:[
"Those of you familiar with the Elm style may recognize that the updated --explain messages draw heavy inspiration from the Elm approach." Turner, Sophia June. https://blog.rust-lang.org/2016/08/10/Shape-of-errors-to-come/[Shape of errors to come]. Rust-Lang, 2016-08-10.
]

Here is an example written in Zig.

```zig
enum LexemeType {
    Whitespace,
    Ident,
    Comment,
    LParen,
    RParen,
    Number,
}

struct Lexeme {
    lexeme_type: LexemeType,
    start: u32, // Mapping to original text
    close: u32, // Mapping to original text
}

// []const u8 is Zig's string
fn lexer(input: []const u8, output: *std.ArrayList(Lexeme)) {
    enum States {
        Whitespace,
        Alphabet,
        Function,
        Number,
    }

    let mut state = States.Whitespace;
    let mut cursor: u32 = 0;

    for (0.., input) |i, ch| {
        switch (state) {
            .Whitespace => switch (ch) {
                'a'..'z' => {
                    state = .Alphabet;
                    cursor = i;
                }
                _ => {}
            },
            .Alphabet => switch (ch) {
                ' ' => {
                    state = .Whitespace;
                    output.append(Lexeme {
                        .lexeme_type = .Ident,
                        .start       = cursor,
                        .close       = i,
                    });
                    cursor = i;
                }
                _ => {}
            },
        }
    }
}
```

## Table-Driven Lexer

I've always heard that you could in theory , that table-driven lexing is possible.{wj}footnote:[
Lexing, probably with no exceptions, can be solved as a Deterministic Finite Automata (DFA), which means they can always be solved by state transition table (by definition of a DFA).
So you can always write a lexer as table-driven, but the question was always how exactly.
]

Essentially, we are trying to reduce from handmade:

```zig
    var state = ...;
    for (input) |ch| {
        state, var output = switch (state) {
             .state1 => switch (ch) {
                 ' ' => .{ state2, .token1 }
                 ...
             },
             .state2 => switch (ch) {
                 'a' => .{ state2, .token1 }
                 ...
             },
             .state3 => switch (ch) { ... },
        }
    }
```

to the following:

```zig
    var state = ...;
    for (input) |ch| {
        state, var output = state_transition[state][ch];
    }
```

For single-threaded code, the biggest hits to performance come from cache misses and branch mispredictions.
Table driving the lexing process removes all branches, so we never have a branch misprediction.
The transition table, depending on the number of states you have, is small enough to fit into the L1 cache (~64 KB)  or L2 cache (~512 KB).
Thus, table-driven lexing is essentially optimal.

For my implementation of a table-driven lexer, one of which you can find link:https://github.com/yueleshia/eyg/blob/main/src/s1_lexer.zig[in my implementation of EYG], I had three key insights:

* Using the same switch statements in the handmade case to generate the transition table
* An insight for handling errors/variable output
* Mapping into character classes (this is strictly an optimization)


### Making the Transition Table

When you think about the handmade case, it essentially boils down to an `if state` and `if character`, update the state and emit a token.
This sounds suspiciously like 2D table:

```zig
    for (input) |x| {
        var new_state, var output = transition_table[state][char];`.
    }
```

But how does one create this transition table?
One could hand-code a table, but that is very unmaintainable.

I think best solution is to have the switch statements that you would have written in the handmade lexer tweaked to instead generate a transition table.
For most languages, you would have complicate your build process to precompute this table, but for languages like Zig (comptime), Go (embed), Rust (`const fn`), this can be done at compile-time.
Here's what it looks like:

```zig
const std = @import("std");

const State = enum { alphabet, whitespace };
const Control = enum { old, new, err }; // Explained later
const Value = enum { ident, text };
const Error = error { SomeError };
// Evaluated at compile time
const transition_table = blk: {
    const states = std.meta.tags(State);
    var ret: [states.len][256]struct { State, Control, Error!Value } = undefined;

    for (states) |state| {
        for (0..256) |ch| {
            ret[@intFromEnum(state)][ch] = switch (state) {
                .whitespace => switch (ch) {
                    'a'..'z' => .{ .alphabet, .new, .ident },
                    ' ' => .{ .whitespace, .old, .text }
                    else => .{ .whitespace, .err, err.SomeError },
                },
                .ident => switch (ch) {
                    'a'..'z' => { .ident, .old, .ident },
                    ' ' => .{ .whitespace, .new, .text }
                    else => .{ .whitespace, .err, err.SomeError },
                }
            };
        }
    }
    break :blk ret;
};

fn lex(input: []const u8, output: *std.ArrayList(Value)) {
    var state = State.Whitespace;
    for (input) |ch| {
        state, const _, var result = transition_table[state][ch];`.
        const token = result catch {
            @panic("Handle errors later");
        };
        output.append(token);
    }
}
```

Aside from the error handling, the body of lex is branchless.

### Grouping Output and Emitting Errors

With our current implementation of `lex`, there is a single output value for every input character.
But what if you want to group letters into a single identifier?
For example, I would want to turn 'var lexeme' into three tokens (`var`, `{nbsp}`, `lexeme`) rather than ten tokens (`v`, `a`, `r`, ...).

I solved this issue by adding the Control enum as part of the transition table.

```zig
const Control = enum(u8) {
    old = 0, // Extend the previous token
    new = 1, // Output a new token
    err = 2, // Output an error
};
fn lex(input: []const u8, output: []Value) {
    var state = State.Whitespace;
    var token_cursor = 0;
    var output_cursor = 0;
    for (0.., input) |i, ch| {
        state, const control, var result = transition_table[state][ch];`.
        // Add 1 if .new
        output_cursor += [_]u8{0, 1, 0}[@intFromEnum(control)];
        // Keep the same if .old else the next token starts at i
        token_cursor = [_]u8{token_cursor, i, i}; // Use for mapping a token back to source

        output[output_cursor] = result catch {
            @panic("Handle errors later");
        };
    }
}
```

Now all branches are removed except for the error.
One solution is to reserve the 0-th index of the output for errors.

```zig
fn lex(input: []const u8, output: []Error!Value) Error!void {
    for (0.., input) |i, ch| {
        state, const control, var result = transition_table[state][ch];`.
        output_cursor += [_]u8{0, 1, 0}[@intFromEnum(control)];
        token_cursor = [_]u8{token_cursor, i, i};

        const idx = [_]u8{output_cursor, output_cursor, 0}[control];
        output[idx] = result;
    }

    output[0] catch |err| {
        return err;
    };
    return output[1..];
}
```

With some cleanup, that's essentially what the final body of lex looks like.

### Optimization: Character Classes

My link:https://github.com/yueleshia/eyg/blob/main/src/s1_lexer.zig[lexer] for EYG, a low syntax language, has 16 states.
This means the transition table is 256 * 16 = 4096 bytes.
This well within the L1 cache, but I can easily a lexer requiring a lot more states.
Sean Barrett suggests using character equivalence classes to reduce this size.{wj}footnote:[Barrett, Sean. https://www.nothings.org/computer/lexing.html[Some Strategies For Fast Lexical Analysis when Parsing Programming Languages]. Sean Barrett, 2015-05-01.]

First we match characters into classes like 'a_to_z', 'newline', etc.
Then your for loop looks something like this

```zig
const Class = enum {
    a_to_z,
    number,
}
// Evaluated at compile time
const equivalence_class: [256]Class = blk: {
    var ret: [256]Class = undefined;
    for (0..10) |i| ret['0' + i] = .number;
    for (0..26) |i| ret['A' + i] = .a_to_z;
    for (0..26) |i| ret['a' + i] = .a_to_z;
    break :blk ret;
};

fn lex(input: []const u8, output: []Error!Value) Error!void {
    for (0.., input) |i, ch| {
        const class = equivalence_class[ch]; // NEW
        state, const control, var result = transition_table[state][class]; // MODIFIED
        output_cursor += [_]u8{0, 1, 0}[@intFromEnum(control)];
        token_cursor = [_]u8{token_cursor, i, i};

        const idx = [_]u8{output_cursor, output_cursor, 0}[control];
        output[idx] = result;
    }

    output[0] catch |err| {
        return err;
    };
    return output[1..];
}
```

# Closing Thoughts

In JAI, of the entire compilation process, lexing and parsing is roughly less than 5% of the total compiler time.{wj}footnote:[Blow, Jonathan. link:https://youtu.be/MnctEW1oL-E?t=1362[Discussion: Making Programming Language Parsers, etc. (Q&A is in a separate video]. 22:42-23:02. YouTube, Jonathan Blow, 2020-03-30.]
Although, The speed of a compiler does I imagine lexing is probably at most 1% of compilation time.
So one might argue that optimizing this part of a compiler is not that fruitful.

But surprisingly, from writing my own table-driven lexer, I actually found that it was a lot easier to iterate on the table-driven implementation.
Because the implementation is so simple, assuming your programming language has a pleasant way to compile-time generate tables, it is extremely to easy to read and make modifications.
The only difficult part of the implementation was solving how to branchlessly output different sized tokens (e.g. "const" vs "for") or errors.
But, I no longer have to modify that part of the code, even when evolving my target programming language.
And this was good practice for branchless programming.

Now that I have been through it once, I think I will always program lexer as table-driven.
Some error messages that I would have emitted in the lexer were moved to the parser, because of the lack of leeway in the table-driven implementation, but for the most part, other than that, it has not been limiting.
Rather, it's been more pleasant to design with, so surprisingly, there have been no down sides, unlike most optimizations.
Assuming you have a good compile-/build-time table generation programming language.

# See also

* I have already reference this article, but many thanks to https://www.nothings.org/computer/lexing.html[this article] by Sean Barrett (RAD Game Tools game programmer), nothings, 2015-05-01.
Barrett walks the reader through to table-driven lexing via a set of optimization decisions.
Table-driven parsing finally clicked for me after reading this.

* And https://www.youtube.com/watch?v=rq1DRuB9p7w[here] is a discussion between Barett and Jonathan Blow (Thekla game designer/programmer) on Barrett's C Compiler.

* Also see the weekly compiler discussions hosted by link:https://www.youtube.com/@compilers[Cliff Click] (author of Sea of Nodes).
