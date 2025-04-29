# Porting a PICO-8 game to PIGO8

PICO-8 is a wonderful fantasy console with its own Lua-based game engine, but once your ideas outgrow the 128×128 constraint you may want to move to a general-purpose language.

Go is a simple, fast, modern language – and thanks to the pigo8 library you can actually port PICO-8 code almost line-for-line. 

In this guide we’ll walk step-by-step through taking the **NerdyTeachers** [“Animate Multiple Sprites”](https://nerdyteachers.com/PICO-8/Game_Mechanics/4) PICO-8 tutorial and rewriting it in Go. 

We’ll start by setting up a project, extracting the sprite sheet with the parsepico tool, and then porting the Lua tables, animation timing, and update/draw loops into Go structs and methods.

Even if you’ve never used Go, we’ll explain how things like static typing and methods work along the way.

By the end you’ll have a running Go program with the same sprite animation logic which is essential for making games.