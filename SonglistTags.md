# Songlist Tags

The following options can be specified in a songlist file.
All tags are specified with the format `{<tag>}`, generally on its own line in the file.
All tags are optional.

## Fixed Order
`{fixed_order}`

If used, the songs will not be re-ordered to fit on the page, rather they will be used in the order given in the songlist file.

## Index Sections
`{index_use_sections}`

If used, the songs will be arranged according to their section, and the index page will also be divided into sections.

## Index Choruses
`{index_use_chorus}`

If used, the index page will include the first lines of the choruses of each song (chorus lines are shown in the index in all caps).

## Index Position
`{index_position: <pos>}`

Determines the position of the index page.  
Valid options are:

`start`

`end`

`none`

Example:
`{index_position: start}`