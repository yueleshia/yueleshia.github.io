{
  "title":   "Password Management Done Right",
  "date":    "2025-12-07",
  "updated": "2025-12-07",
  "layout":  "post.shtml",
  "tags":    ["security", "local-first"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

!{#run: BUILD=build tetra run file % >index.smd }

!{#
@TODO:
* Improve introduction
* Test `gopass sync`
* pass grep
* conclusion
* Web of trust
#}

# Introduction

If you don't already, you should be using a password manager.[^dont-forget-two-factor]
You remember just one (more complicated password) to your password manager that unlocks the passwords for all other accounts
Unlike the rest of security, it is both more convenient and more secure.

For most people, nothing quite beats the convenience of cloud-based password managers.[^cloud-based-password-manager]
But what if you could not rely on any external services, passwords all on your local computer?
Well, here's how featuring [gopass](https://github.com/gopasspw/gopass).

[^dont-forget-two-factor]: Password managers do not absolve you of need for two-factor authentication. Passwords fundamentally are a static knowledge check. Two-factor protects against guessing/stealing your knowledge by requiring something hard-to-copy (i.e. physical with a unique fingerprint) to corroborate your identity.

[^cloud-based-password-manager]: If you do choose to go this route, look for 1) self-hostable solutions (even if you do not self-host) for more trust-worthy code and 2) local-first in case you are without internet.

This post's threat model is someone stealing our passwords via the internet.
And if someone gets physical access to your computer, you can still master-password protect your private key.

# The Pass/Gopass User Experience

The [pass](https://www.passwordstore.org/) or `password store` is a CLI tool that operates entirely locally.
It has terrible searchability, but it follows the [UNIX philosophy](https://en.wikipedia.org/wiki/Unix_philosophy), so it interoperates with other software.
Chiefly among which, pass uses [GPG](https://en.wikipedia.org/wiki/GNU_Privacy_Guard) which is the gold standard for encryption.
You can share the encrypted `.gpg` files to wherever and to whomever you want, yes, even to the strangers on the internet.

Pass maintains a folder on your computer called a 'password store'.
Each password you add with `pass generate <name>` or `pass edit <name>` is just `.gpg` file in said password store.
In fact, pass is just couple-hundred-ish line [bash script](https://git.zx2c4.com/password-store/tree/src/password-store.sh) that essentially automates `gpg --encrypt` and `gpg --decrypt`.

Backing up is done by pushing to a git repo or any other file syncing solution.
I use [syncthing](https://en.wikipedia.org/wiki/Syncthing).

So why `gopass` and not `pass`?
Pass out of the box already supports multiple people.[^sharing-with-pass]
What gopass does for you is automate the sharing of public keys by adding exporting all relevant public keys to the password store in .
You will still have to do the `gpg --import` yourself, but everything else is automated via the `gopass recipients add` command.

[^sharing-with-pass]: Assuming you handle importing several public keys yourself, pass actually supports multiple public keys out of the box.
Simply run `pass init key1 key2`, and it will re-encrypt all your passwords with all keys you specified.

So day to day, what I use are:

* `pass generate path/to/secret` to generate a random password
* `pass show` to show or tab complete my password list (`gopass ls`)
* `pass show path/to/secret` to display my password
* `pass show -c path/to/secret` to copy my password
* `pass edit path/to/secret` to edit my password with whatever is set as $EDITOR (e.g. vim)
* `gopass recipients add` to add and re-encrypt with a new public key
* `gopass sync` if you use git. I don't use this command because I use syncthing.


# How the Encryption Works

The only way to trust a security solution is to understand how it works. (And to know what you are protecting against, i.e. have a threat model)
If you know how asymmetric cryptography works conceptually already, then skip to the [Sharing with gopass](#sharing-with-gopass).
Now, we will build that knowledge from the ground up.

## Cryptography Basics

__Encryption__ and __decryption__ are two parts of cryptography that, in practical terms, are the defined set of steps (aka. algorithm or protocol) to mix and recover a message with a secret so that only those with the secret can read the message.
For example, to transmit the message `hello`, the sender could add one to each letter to get `ifmmp` and the receiver would subtract one from each letter.
This is [Caesar Cipher](https://en.wikipedia.org/wiki/Caesar_cipher) (the classic), with a secret of `+1`.

__Symmetric__ encryption and decryption is are the set of algorithms where the secret is a password, text that both parties know.
__Asymmetric__ encryption and decryption, aka. public-key encryption, is when one secret key (the private key) is used for decryption but a second, non-secret key (the public key) is used for encryption.
The private key never leaves your PC and is never communicated to anyone, where as the public key is always communicated to everyone.
Thus the ingredients for each are:

|                       |Agreed algorithm|Message|Secret|Public|
|---|---|---|---|---|
|Symmetric de/encryption| Y              | Y     | Y    | N/A |
|Asymmetric encryption   | Y              |       |      | Y   |
|Asymmetric decryption   | Y              | Y     | Y    |     |

You can ignore the 'agreed algorithm' column.
Typically, it is difficult to hide the algorithm you are using if you need to talk to someone other than yourself.
The premise of this post is sharing (probably through the internet), and the internet is networking hardware + a long list of commonly agreed communication protocols.

## Asymmetric Cryptography

Like most others explanations, I will hand wave the definition of public/private keys.
Just imagine that they are prime numbers, and that it is computationally difficult to .

The guarantee of asymmetric cryptography is it is difficult to (billions of years)[^brute-force-rsa] derive the public from the private; but the reverse is often not true.
Other than this guarantee, there is not much else that distinguishes the two.[^signing-as-encryption]

[^brute-force-rsa]: @TODO

[^signing-as-encryption]: There's nothing stopping you from also encrypting with a private key.
In fact, this exactly what the emergent [`signing`](https://en.wikipedia.org/wiki/Digital_signature) capability is.
Signatures are meant to prove identity.
Unlike encryption, the message is known (so algorithm, message, and public are known by all parties).
Signing encrypts a hash of the message, which can be decrypted with the public key and validated against the hash of the message.

The secret key, effectively your one-way password, is only accessible as a file on your computer.
Sending your public key is effectively registering your device on a target.
Here are some illustrative use cases:

* Pass: Stores passwords as text documents that are encrypted by the public key. Adding a key registers a user as someone who can decrypt the password store.

* SSH: One key pair is associated with for one user's one device. A public keys is thus effectively your 'login', and has to be added to a list of authorized users.

* [Wish](https://github.com/charmbracelet/wish): A go library for managing user as password-less logins with just public keys. This is the same concept as logging in with SSH on GitHub.

The cool thing about public key cryptography is that you publish the encrypted files publicly, as long as you never share you private key (effectively impossible to brute force).
Your private key never touches the internet, and you (*should*) never copy your private keys.
But if your key only ever exists on your one device, how does one sync to another or share passwords?

Files encrypted by GPG end with `.gpg` by convention.
These are actually lists of encrypted `packets` of data.
So `gpg` supports multiple public keys encrypting a single message, and just joins them together.
You can verify this yourself.

```sh
#gopass generate a # will create the file ~/.password-store/a.gpg
gpg --list-packets

#:pubkey enc packet ...
#:pubkey enc packet ...
```

[]($section.id('sharing-with-gopass'))
## Sharing with gopass (or pass)

Let's say you have already have a key pair, have initialised your password store, and have secrets. E.g.

```sh
gpg --full-generate-key
# (1) RSA and RSA
# 4096 bits

gopass init # Stores which key pair you want to use for ~/.password_store
gopass generate somepass # Create a password named 'somepass'
```

Now let's run through the mechanics of how to share/backup your passwords.

* A second person wants to join your password store. Let's call them FriendB.
* FriendA obtains a local copy of your password store by running `gopass init`.
These secrets have one packet.
This will add their public key to their local copy.
This add their public key to `~/.password_store/.public-key`.
* You sync with FriendA to copy over their public key.
* You re-encrypt all secrets by running `gopass recipients add`.
* You sync with FriendA to overwrite all of all their local secrets, now with two packets.

## Web of Trust

The last concept that is important, but irrelevant to the pass use case, is [web of trust](https://en.wikipedia.org/wiki/Web_of_trust).
Skip now, lest you obtain world knowledge.

The web of trust is solving around associating public keys with people.
For example, you making the first contact to a person and are downloading their public key off a sketchy website, or you have a receiving the public key from a friend over the internet.

In concrete steps, the web of trust means that individuals are asserting that a public key is indeed associated with the person (email and name) within the public key.
Letting others know you trust a public key means signing a public key, and uploading your public key (i.e. identity) to key servers (which are decentralized, for example [MIT's key server](https://pgp.mit.edu/)).

You assign trust levels (not a cryptographic concept) to other public keys to indicate how much you trust said person to verify other's public keys.
You do this for several public keys, and then this gives an ad-hoc indication for likely a random public key is to truly be associated with the person in question.
And by participating, you are building the web of trust.

Ironically, this is the most important part of GPG (but irrelevant to our use case).
If you have ever heard of the [fourteen people or seven keys](https://www.icann.org/en/blogs/details/the-problem-with-the-seven-keys-13-2-2017-en) that control all of the internet (DNS), all of DNS and SSL are essentially the same exact problem.
In [FUTO ID](https://docs.polycentric.io/futo-id/), a modern take on decentralized identity, claims are the same concept.


# Comparison of Cloud to GoPass

|   |Cloud (most services) |GoPass|
|---|---|---|
|Single password unlock | Y         | Y, GPG or GPG with password
|Device syncing         | Y         | Y, with any file syncing method (git, syncthing, dropbox, etc.)
|Family sharing         | Y         | Y, use different stores with varying recipients
|Team sharing           | ?         | Y
|Local-first            | ?         | Y
|Offline-usable         | I certainly hope so | Y
|Self-hostable          | Y         | N/A, as it is entirely offline
|Browser integration    | Y         | [Y](https://en.wikipedia.org/wiki/Digital_signature)
|Mobile app             | Y         | Not-first class, but [possible](https://github.com/gopasspw/gopass/issues/571)
|CLI friendly           | ?         | Y
|Desktop GUI            | Often browser first | [Y](https://github.com/codecentric/gopass-ui)
|Immune to online brute-force | N, can brute force your login | Y, even if you publish your encrypted `.gpg` passwords online, you never publish secret key (and GPG is effectively non-brute-forceable).

# See also

* [Intro to pass](https://www.youtube.com/watch?v=FhwsfH2TpFA) by Dreams of Code for more advanced use cases.
* [Temo workflow demo](https://www.youtube.com/watch?v=EB9cW9RjiSs&t=1076) of gopass by Andrew Topin. The only video I found demoing this features
* https://woile.github.io/gopass-cheat-sheet/
* @TODO: Something to explain best practice around GPG key rotation, air gapping, expiry, web of trust, etc. to have good long-term security
