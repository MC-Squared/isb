# Song Tags

Below is a list of currently supported tags in .song files. All tags have the form {tag}. All tags are optional.

## Title
`{title: <title>}`

This is used to set the title of the Song, which will appear on the index page. If no title is specified, then the first line will be used by default.

Example:
`{title: I want to be filled with the Triune God}`

## Section
`{section: <section>}`

This is used to specify the section of the song. If used, the sections will then affect the order of the songs and will be shown on the index page.

Example:
`{section: Worship of the Father}`

## No Number
`{no_number}`

This is used to indicate that a stanza should not display a stanza number when printed (useful for some short songs or two-part songs).

Example: 
```
{no_number}
Therefore with joy...
```

## Comments
`{comments: <comment>}`

Used to provide text that is not sung, can be placed before or after any stanza. Multiple comments tags can be used, each one will be placed on its own line.

Example: 
```
{comments: Hymns, #1008}
{comments: Capo 1}
```

## Chorus
```
{start_of_chorus}  
{end_of_chorus}
```

Used to indicate that the lines in between are the chorus. Choruses are treated the same as stanzas except they do not display a stanza number and they are indented when printed.

Example:

```
...
That he may [A]live for[D]e’er.

{start_of_chorus}
God is in [G]Christ to be my sup[D]ply,
God as the Spirit nourisheth [G]me;
If upon Christ in spirit I [C]feed,
Filled with His [D]life I’ll [G]be.
{end_of_chorus}

The tree the glorious Christ does show
...
```

## Echo
`{echo: <echo text>}`

Used to indicate that part of a line is sung as an echo (or is optional).
When printed the echo text will be slightly lighter in colour (i.e. grey).
Note: Echo tags can be inline (as in example below) or on their own line.
Only one echo tag can be used per-line, and no text should come after the tag on the same line.

Example:
```
And the bread {echo: and the bread}
```