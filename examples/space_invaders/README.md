# Space Invaders

A classic Space Invaders game implemented using PIGO8, inspired by the 1978 arcade hit.

## Controls

- **Left/Right Arrow Keys**: Move spaceship
- **A Button**: Shoot
- **A Button (Game Over)**: Restart game

## Features

- Classic Space Invaders gameplay
- Multiple alien types with different point values
- Score tracking
- Lives system
- Game over and restart functionality

## How to Run

```bash
cd examples/space_invaders
go run .
```

## Gameplay

- Destroy all alien invaders before they reach the bottom of the screen
- Each alien shot gives you 10 points
- You have 3 lives - if an alien bullet hits you or aliens reach the bottom, you lose a life
- Game ends when you run out of lives
- Clear all aliens to advance to the next level

## Technical Details

- Built using PIGO8's rendering and input systems
- Implements simple collision detection
- Uses sprites for the player and aliens
- Features simple AI for alien movement and shooting patterns
