# Parse PICO-8 sprites

Create a new folder, where you will place your project's code.

```bash=
mkdir mygame; cd mygame
```

## Fetch your p8 game into the directory

Copy from [NerdyTeachers .p8 cartridge](https://nerdyteachers.com/PICO-8/resources/img/tutorials/animateSprite/animate_sprites.p8.png) (`p8` text file) into this folder.

PICO-8 carts come in two formats: the text-based `.p8` format and the `“.p8.png”` format which hides the code/data inside a PNG image. 

The **text** `.p8` file contains the Lua source and data sections in plain text, while the `.p8.png` is a 128×128 image containing the same data (along with a screenshot). See:

![Game](https://nerdyteachers.com/PICO-8/resources/img/tutorials/animateSprite/animate_sprites.p8.png)

You can load the .p8 file in PICO-8 or view it in a text editor to see the Lua code and sprite data.

## Get the graphics

We want the sprite graphics from this cart. Namely, we mean these sprites:

![PICO8Sprites](pico8_sprites.png)

The simplest approach (_but we won't do this_) is to export the sprite sheet PNG from PICO-8 (by typing `export sprites.png` in the console). This would mean we need to write code to improt this PNG sprisheet.png and slice it in the code, not to mention there is no way to go around PICO-8 flag's configuration for sprites. To avoid such a thing, we will use another tool. One such tool is [parsepico](https://github.com/drpaneas/parsepico) which can **read** a `.p8` file and spit out the sprite images, maps, along with `JSON`s that have all the required metadata for PIGO8.

For example:

Fetch your `*.p8` game and place it into that folder:

```bash
cp $PICO8/carts/animate_sprites.p8 .
```

Make sure the PICO-8 file is text file, and not PNG (e.g. `p8.png` is not supported).

```bash
# This is supported:
% file animate_sprites.p8 
animate_sprites.p8: ASCII text


# This is not supported (you have to open PICO-8 and save it as *.p8)
% file animate_sprites.p8.png 
animate_sprites.p8.png: PNG image data, 160 x 205, 8-bit/color RGBA, non-interlaced
```

Great so, the next step is to extract the sprites for this game.
To do that, we will use a tool called [parsepico].
You can either fetch it from the release page in Github, or use Go packaging mechanism to install it directly to your system.

```bash
$ go install github.com/drpaneas/parsepico@latest
```

```bash
# Expected Output
go: downloading github.com/drpaneas/parsepico v1.0.6
```

If you follow this way, Go will download and save it at `$GOPATH/bin`.
You can verify this:

```bash
file `go env GOPATH`/bin/parsepico

# Output:
/Users/pgeorgia/gocode/bin/parsepico: Mach-O 64-bit executable arm64
```

To use it either call it from this location or add it to your `PATH`:

```bash
$ export PATH="`go env GOPATH`/bin:$PATH"
```

Now, you should be able to run this:

```bash
$ parsepico --help
```

```bash
# Expected Output
Usage of parsepico:
  -3	Include dual-purpose section 3 (sprites 128..191)
  -4	Include dual-purpose section 4 (sprites 192..255)
  -cart string
    	Path to the PICO-8 cartridge file (.p8)
  -clean
    	Remove old sprites directory, map.png, spritesheet.png if they exist
```

Note: 
> Using Go package manager is not required.
> You can always fetch the executable/binary directly from the release page of [parsepico] for your OS and architecture.

So now we confirmed we have the `*.p8` game and `parsepico` installed, we can extract the graphics from this game:

```bash
$ parsepico --cart=animate_sprites.p8
```

This will parse the PICO-8 cart and generate several useful things.
From all of these, we are only interested in `spritesheet.json` which will be using in our PIGO8 game.

```bash
# Expected Output
No __map__ section found. Skipping map processing.
Saved 4 sections into 'sprites' folder.
Created spritesheet.png with 4 sections.
Successfully generated spritesheet.json # we will need only this
Successfully created individual sprite PNGs
```

That said, feel free to delete the rest of the generated files, to free some disk space:

```bash
$ rm -r sprites
$ rm spritesheet.png 
```

Ok, so no we are ready to start writing Go code!