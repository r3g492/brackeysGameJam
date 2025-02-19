package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	gameObjects          = make(map[int]GameObject)
	nextGameObjectId int = 0
)

func main() {
	rl.InitWindow(1600, 900, "Nothing can go wrong...")
	defer rl.CloseWindow()
	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	diamondTexture2D := rl.LoadTexture("resources/diamond.png")
	diamond := Diamond{
		id:        nextGameObjectId,
		texture:   diamondTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 200, Height: 200},
		position:  rl.Vector2{X: 200, Y: 300},
		color:     rl.Blue,
	}
	nextGameObjectId++
	gameObjects[diamond.id] = &diamond

	for !rl.WindowShouldClose() {

		staticId := -1
		if rl.IsKeyDown(rl.KeyS) {
			newDiamond := Diamond{
				id:        staticId,
				texture:   diamondTexture2D,
				sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 200, Height: 200},
				position:  rl.Vector2{X: 500, Y: 500},
				color:     rl.Blue,
			}
			nextGameObjectId++
			gameObjects[staticId] = &newDiamond
		} else {
			delete(gameObjects, staticId)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.DrawText("Congrats! You created your first window!", 190, 200, 20, rl.Black)

		DrawGameObjects()
		rl.EndDrawing()
	}
}

func DrawGameObjects() {
	for _, obj := range gameObjects {
		obj.Draw()
	}
}

type GameObject interface {
	Draw()
	GameObjectId() int
}

type Diamond struct {
	id        int
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

func (d *Diamond) GameObjectId() int {
	return d.id
}
