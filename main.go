package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"time"
)

var (
	gameObjects          = make(map[int]GameObject)
	nextGameObjectId int = 0
)

// i = v / r game

func main() {
	display := rl.GetCurrentMonitor()
	userMonitorWidth := rl.GetMonitorWidth(display)
	userMonitorHeight := rl.GetMonitorHeight(display)
	screenWidth := int32(userMonitorWidth)
	screenHeight := int32(userMonitorHeight)
	rl.InitWindow(screenWidth, screenHeight, "Electric.")
	rl.MaximizeWindow()
	if !rl.IsWindowFullscreen() {
		rl.ToggleFullscreen()
	}
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	// start menu
	buttonTexture2D := rl.LoadTexture("resources/button.png")
	startButtonScreen(buttonTexture2D, display)
	// start menu end

	// https://pixabay.com/sound-effects/female-vocal-321-countdown-240912/
	countdownSound := rl.LoadSound("resources/female-vocal-321-countdown-240912.mp3")
	// count down screen
	countdown(countdownSound, display)
	// count down end

	gameTimer := Timer{}
	gameTimer.Init()

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
		rl.DrawText(
			fmt.Sprintf(
				"Time elapsed: %.0f s",
				time.Since(gameTimer.gameInitTime).Seconds(),
			),
			190,
			200,
			20,
			rl.Black,
		)

		if hasWon() {
			calculateScore()
			showScore()
		}

		staticId := -1
		if rl.IsKeyDown(rl.KeyS) {
			newDiamond := Diamond{
				id:        staticId,
				texture:   diamondTexture2D,
				sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 200, Height: 200},
				position:  rl.Vector2{X: 500, Y: 500},
				color:     rl.Blue,
			}
			gameObjects[staticId] = &newDiamond
		} else {
			delete(gameObjects, staticId)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		DrawGameObjects()
		rl.EndDrawing()
	}
}

func startButtonScreen(buttonTexture2D rl.Texture2D, display int) {
	button := Button{
		id:        -1,
		texture:   buttonTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 220, Height: 100},
		position: rl.Vector2{
			X: float32(rl.GetMonitorWidth(display))/2 - 220/2,
			Y: float32(rl.GetMonitorHeight(display))/2 - 220/2,
		},
		color:  rl.White,
		status: 0,
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		mousePoint := rl.GetMousePosition()
		if button.CheckInput(mousePoint) {
			break
		}
		button.Draw()
		rl.EndDrawing()
	}
}

func countdown(countdownSound rl.Sound, display int) {
	beginTimer := Timer{}
	beginTimer.Init()
	rl.PlaySound(countdownSound)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		secondsLeft := 1 - time.Since(beginTimer.gameInitTime).Seconds()

		rl.DrawText(
			fmt.Sprintf(
				"Be Ready!",
			),
			int32(rl.GetMonitorWidth(display)/2-200),
			int32(rl.GetMonitorHeight(display)/2-300),
			300,
			rl.Black,
		)
		if secondsLeft <= 0 {
			break
		}
		rl.EndDrawing()
	}
}

func hasWon() bool {
	return false
}

func calculateScore() {

}

func showScore() {

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

type Timer struct {
	gameInitTime time.Time
}

func (t *Timer) Init() {
	t.gameInitTime = time.Now()
}

type Button struct {
	id               int
	texture          rl.Texture2D
	sourceRec        rl.Rectangle
	position         rl.Vector2
	color            rl.Color
	status           int
	buttonClickSound rl.Sound
}

func (b *Button) Draw() {
	if b.status == 0 {
		b.sourceRec.Y = 0
	} else if b.status == 1 {
		b.sourceRec.Y = 110
	} else if b.status == 2 {
		b.sourceRec.Y = 220
	}
	rl.DrawTextureRec(
		b.texture,
		b.sourceRec,
		b.position,
		b.color,
	)
}

func (b *Button) CheckInput(
	mousePosition rl.Vector2,
) bool {
	if mousePosition.X >= b.position.X &&
		mousePosition.X <= b.sourceRec.Width+b.position.X &&
		mousePosition.Y >= b.position.Y &&
		mousePosition.Y <= b.sourceRec.Height+b.position.Y &&
		rl.IsMouseButtonDown(rl.MouseLeftButton) {
		b.status = 1
	} else {
		b.status = 0
	}

	if mousePosition.X >= b.position.X &&
		mousePosition.X <= b.sourceRec.Width+b.position.X &&
		mousePosition.Y >= b.position.Y &&
		mousePosition.Y <= b.sourceRec.Height+b.position.Y &&
		rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		b.status = 2
		return true
	}
	return false
}

func (b *Button) GameObjectId() int {
	return b.id
}
