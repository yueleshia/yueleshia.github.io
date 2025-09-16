{
  "title":   "Archive: Proposing UTF-24",
  "date":    "2007-01-17",
  "updated": "2025-09-21",
  "author":  "Ruszlan Gaszanov",
  "layout":  "post.shtml",
  "tags":    ["archive", "text"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

Why would we need a new UTF?
Well, all currently available UTFs have their advantages and flaws and I see another possible UTF which might have quite a few nice features no existing UTF offers in bundle.
In fact, the form of encoding I'm about to propose seems so obvious to me, that I only wonder why no one has seriously proposed it before.
But let's see strong and weak points of the existing UTF encoding schemes:

# UTF-32

## Advantages:

1. Fixed length code units
1. Convenient for internal processing on 32/64-bit platforms
1. All code units can be interpreted as binary representation of Unicode code points scalar values

## Disadvantages:
1. 11 unused bits per code unit is much too wasteful for long-term storage and interchange.
2. Explicit agreement of byte order (out-of-band or by BOM) required.
3. Dropped or insertet octets in the middle of data stream are likely to corrupt the remainder of the stream.
4. Despite 11+ACE- spare bits per code unit, no attempt has been made to design some mechanism of error and byte order detection (which again proves that this format was designed mainly for internal processing)
5. Incompatible with many legacy text-processing tools and protocols (though that may be considered a +ACo-good+ACo- thing in some cases)

# UTF-16

## Advantages:
1. Fixed length code units as long as restricted to BMP.
2. Most common characters take half the space they would in UTF-32, while the rest take exactly as much.
3. Convenient for internal processing on 32/64-bit platforms as long as only BMP characters are used.
4. Most compact UTF for East-Asian texts.

## Disadvantages:
1. Variable-length code units.
2. Explicit agreement of byte order (out-of-band or by BOM) required.
3. Dropped or insertet octets in the middle of data stream are likely to corrupt the
   remainder of the stream.
4. Still incompatible with many legacy text-processing tools and protocols.

# UTF-8

## Advantages:
1. Takes less space for most texts then any other UTF (on the average, slightly over 1 octet per character for Latin-based scripts and slightly under 2 octets per character for most other scripts in current use).
1. No agreement of byte order (out-of-band or by BOM) required.
1. Has internal resync mechanism, so a dropped or inserted octet in multibyte sequence
   would only corrupt one character.
1. ASCII-transparent.
1. Compatible with most legacy text-processing tools and protocols (though that may be considered a +ACo-bad+ACo- thing in some cases).

## Disadvantages:
1. Variable-length code units.
1. Requires more processing overhead compared to UTF-16/32.
1. Still incompatible with some older 7-bit protocols.

# UTF-7
## Advantages:
1. Compatible with older 7-bit protocols and more compact then Base64 encoded UTF-8/16.

## Disadvantages:
1. Impractical for any other purpose then providing 7-bit transparency.

#

The encoding scheme I am purposing here is intended to be fixed-length, but, unlike UTF-32 is intended mainly for interchange and storage (though it would still be more efficient for internal processing then UTF-7/8/16).
In order to represent every Unicode code point directly in binary form, minimum 21 bit is needed. However, considering that most data transmission and storage technologies we are dealing with are octet-oriented, it might be more practical to use 24 bits per code point in 3-octet code units.
The most straightforward approach to implement that would be by simply stripping the most significant octet from each UTF-32 code unit. However, I propose a different approach:

    >UTF-32: 00000000 000ustrq ponmlkji hgfedcba
    >UTF-24: zustrqpo ynmlkjih xgfedcba

Hence we have data bits a-u distributed evenly among 3 octets, and 3 unused bits in each octet - x, y and z. The latter can be either all set to 0 to provide 7-bit transparency or be used for byte order and error detection mechanism, like this:

    >x +AD0- 0
    >z +AD0- 1
    >y +AD0- a +AF4- b +AF4- c +AF4- d +AF4- e +AF4- f +AF4- g +AF4- h +AF4- i +AF4- j +AF4- k +AF4- l +AF4- m +AF4- n +AF4- o +AF4- p +AF4- q +AF4- r +AF4- s +AF4- t +AF4- u

So, the +ACI-high+ACI- (most significant) octet would always have 8th bite set to 1, the +ACI-low+ACI- (least significant) octet always set to 0, while the 8th bit of the +ACI-middle+ACI- octet would be the parity of 21 data bits across 3 octets of the code unit.
In fact, because we only have 17 valid planes out of 32 encodable by 5 bits, high octet could only look either like this:

    10xxxxxx

Or like this:

    110000xx

This kind of mechanism would let us both reliably detect byte order without the use of BOM (even if the data consists of only a single code unit) and resync at the next valid code unit in case of dropped / inserted octets.
It would also be possible be possible to start decoding anywhere in the stream without having to count the number of octets from the beginning.
Of course, this proposed encoding scheme would also have its flaws, but hey, there are always tradeoffs. So, let's summarize:

# UTF-24

## Advantages:
1. Fixed length code units.
2. Encoding format is easily detectable for any content, even if mislabeled.
3. Byte order can be reliably detected without the use of BOM, even for single-code-unit data.
4. If octets are dropped / inserted, decoder can resync at next valid code unit.
5. Practical for both internal processing and storage / interchange.
6. Conversion to code point scalar values is more trivial then for UTF-16 surrogate pairs and UTF-7/8 multibyte sequences.
7. 7-bit transparent version can be easily derived.
8. Most compact for texts in archaic scripts.

## Disadvantages:
1. Takes more space then UTF-8/16, except for texts in archaic scripts.
2. Comparing to UTF-32, extra bitwise operations required to convert to code point scalar values.
3. Incompatible with many legacy text-processing tools and protocols.

Any comments?

Ruszl+AOE-n Gaszanov


