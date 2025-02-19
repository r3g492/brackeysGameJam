package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(1600, 900, "Nothing can go wrong...")
	defer rl.CloseWindow()
	rl.InitAudioDevice()

	rl.SetTargetFPS(60)

	var diamondTexture2D rl.Texture2D = rl.LoadTexture("resources/diamond.png")

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.DrawText("Congrats! You created your first window!", 190, 200, 20, rl.Black)

		var diamond = Diamond{
			diamondTexture2D,
			rl.Rectangle{
				0,
				0,
				200,
				200,
			},
			rl.Vector2{
				200,
				300,
			},
			rl.Blue,
		}

		diamond.Draw()

		rl.EndDrawing()
	}
}

type Diamond struct {
	texture   rl.Texture2D
	sourceRec rl.Rectangle
	position  rl.Vector2
	color     rl.Color
}

func (d *Diamond) Draw() {
	rl.DrawTextureRec(
		d.texture,
		d.sourceRec,
		d.position,
		d.color,
	)
}
