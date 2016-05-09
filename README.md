[![wercker status](https://app.wercker.com/status/f19eaf65003b2ab027cc77b8d4be9a51/m "wercker status")](https://app.wercker.com/project/bykey/f19eaf65003b2ab027cc77b8d4be9a51)

# Indigo Song Book

## The overall goal of this project is to create songbooks with:

- Exact chord placement.
- Electronic songbooks.
- Automated index/TOC creation.
- Automatic layout (i.e. shuffling songs around, so that they fit nicely on the page).
- Automatic formatting for verses and choruses.
- Some kind of documentation, so others can pick this up in the future.

## Syntax For Song Files

- Chords are indicated in square brackets, within the text. e.g. [G]I love [Em]God...
- Chorus is marked by {start_of_chorus} and {end_of_chorus}.
- Verses are separated by a blank line.
- Text that is not sung (e.g. marking part 1, part 2) is marked by {comments: text}. 

For a complete list of supported tags see [Song Tags](SongTags.md) page.

## Syntax For Songlist Files

- One filename per line (including ".song" is optional)

For a complete list of supported options, see [Songlist Tags](SonglistTags.md).
