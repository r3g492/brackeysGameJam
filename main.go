package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"strconv"
	"time"
)

var (
	gameObjects          = make(map[int]GameObject)
	nextGameObjectId int = 0
)

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

	buttonTexture2D := rl.LoadTexture("resources/button.png")
	startTexture2D := rl.LoadTexture("resources/start.png")
	if !startButtonScreen(buttonTexture2D, display, startTexture2D) {
		return
	}

	diamondTexture2D := rl.LoadTexture("resources/diamond.png")
	midPointX, midPointY := raylibWindowMidPoint(100, 100)
	player := Player{
		id:            0,
		texture:       diamondTexture2D,
		sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
		position:      rl.Vector2{X: midPointX, Y: midPointY},
		color:         rl.Black,
		movementSpeed: 5,
	}
	gameObjects[0] = &player
	nextGameObjectId = 1

	// https://pixabay.com/music/trap-spinning-head-271171/
	bgm := rl.LoadSound("resources/spinning-head-271171.mp3")
	rl.PlaySound(bgm)
	gunShot := rl.LoadSound("resources/shotgun-03-38220.mp3")
	stageIdx := 0
	stageEnd := 3

	gameTimer := Timer{
		time.Now(),
		rl.Vector2{
			X: float32(rl.GetMonitorWidth(display) / 2),
			Y: float32(0),
		},
	}
	for ; stageIdx < stageEnd; stageIdx++ {
		if !rl.IsSoundPlaying(bgm) {
			rl.PlaySound(bgm)
		}

		// https://pixabay.com/sound-effects/female-vocal-321-countdown-240912/
		countdownSound := rl.LoadSound("resources/female-vocal-321-countdown-240912.mp3")
		countdown(countdownSound, display, fmt.Sprintf("stage %s", strconv.Itoa(stageIdx+1)))

		player.position.X = midPointX
		player.position.Y = midPointY

		// create enemy
		enemy1 := Enemy{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
			position:      rl.Vector2{X: midPointX + 500, Y: midPointY + 500},
			color:         rl.Red,
			movementSpeed: 5,
		}
		gameObjects[nextGameObjectId] = &enemy1
		nextGameObjectId++

		enemy2 := Enemy{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
			position:      rl.Vector2{X: midPointX - 500, Y: midPointY - 500},
			color:         rl.Red,
			movementSpeed: 5,
		}
		gameObjects[nextGameObjectId] = &enemy2
		nextGameObjectId++

		enemy3 := Enemy{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
			position:      rl.Vector2{X: midPointX + 250, Y: midPointY + 250},
			color:         rl.Red,
			movementSpeed: 5,
		}
		gameObjects[nextGameObjectId] = &enemy3
		nextGameObjectId++

		for !rl.WindowShouldClose() {
			if rl.WindowShouldClose() {
				return
			}

			printYourTime(gameTimer, time.Now(), false, display)

			if hasWonStage() {
				if stageIdx >= stageEnd-1 {
					if WinScreen(buttonTexture2D, display, startTexture2D, gameTimer) {
						return
					}
				}
				break
			}

			playerMovement(&player)
			if playerDeathCheck(&player) {
				if gameOverScreen(buttonTexture2D, display, startTexture2D) {
					// restart game
					stageIdx = -1
					gameTimer.Init()
					break
				}
				return
			}

			if rl.IsMouseButtonDown(rl.MouseLeftButton) {
				rl.PlaySound(gunShot)
				isMouseOverEnemy(rl.GetMousePosition())
			}

			rl.BeginDrawing()
			rl.ClearBackground(rl.RayWhite)
			DrawGameObjects()
			rl.EndDrawing()
		}
	}
}

func playerMovement(player *Player) {
	isUpPressed := rl.IsKeyDown(rl.KeyW)
	isLeftPressed := rl.IsKeyDown(rl.KeyA)
	isDownPressed := rl.IsKeyDown(rl.KeyS)
	isRightPressed := rl.IsKeyDown(rl.KeyD)

	var movementPressedKeyCount float32 = 0
	if isUpPressed {
		movementPressedKeyCount++
	}
	if isLeftPressed {
		movementPressedKeyCount++
	}
	if isDownPressed {
		movementPressedKeyCount++
	}
	if isRightPressed {
		movementPressedKeyCount++
	}

	dividedMovementSpeed := player.movementSpeed / movementPressedKeyCount

	if isUpPressed {
		player.position.Y = player.position.Y - dividedMovementSpeed
	}
	if isLeftPressed {
		player.position.X = player.position.X - dividedMovementSpeed
	}
	if isDownPressed {
		player.position.Y = player.position.Y + dividedMovementSpeed
	}
	if isRightPressed {
		player.position.X = player.position.X + dividedMovementSpeed
	}
}

func playerDeathCheck(player *Player) bool {
	playerHitbox := rl.Rectangle{
		X:      player.position.X,
		Y:      player.position.Y,
		Width:  player.sourceRec.Width,
		Height: player.sourceRec.Height,
	}

	for _, obj := range gameObjects {
		if obj.IsEnemy() {
			enemyHitbox := obj.Hitbox()
			if rl.CheckCollisionRecs(playerHitbox, enemyHitbox) {
				return true
			}
		}
	}
	return false
}

func startButtonScreen(
	buttonTexture2D rl.Texture2D,
	display int,
	startTexture2D rl.Texture2D,
) bool {
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
			return true
		}
		rl.DrawTextureRec(
			startTexture2D,
			rl.Rectangle{X: 0, Y: 0, Width: 1600, Height: 900},
			rl.Vector2{X: float32(rl.GetMonitorWidth(display))/2 - 800, Y: float32(rl.GetMonitorHeight(display))/2 - 450},
			rl.Gray,
		)
		button.Draw()
		rl.EndDrawing()
	}
	return false
}

func gameOverScreen(
	buttonTexture2D rl.Texture2D,
	display int,
	startTexture2D rl.Texture2D,
) bool {
	button := Button{
		id:        -1,
		texture:   buttonTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 220, Height: 100},
		position: rl.Vector2{
			X: float32(rl.GetMonitorWidth(display))/2 - 220/2,
			Y: float32(rl.GetMonitorHeight(display))/2 - 220/2,
		},
		color:  rl.Red,
		status: 0,
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		mousePoint := rl.GetMousePosition()
		if button.CheckInput(mousePoint) {
			return true
		}

		rl.DrawTextureRec(
			startTexture2D,
			rl.Rectangle{X: 0, Y: 0, Width: 1600, Height: 900},
			rl.Vector2{X: float32(rl.GetMonitorWidth(display))/2 - 800, Y: float32(rl.GetMonitorHeight(display))/2 - 450},
			rl.Gray,
		)
		rl.DrawText(
			fmt.Sprintf(
				"you died. \n you must be perfect...",
			),
			int32(rl.GetMonitorWidth(display)/2-400),
			int32(rl.GetMonitorHeight(display)/2-400),
			100,
			rl.Red,
		)
		button.Draw()
		rl.EndDrawing()
	}
	return false
}

func WinScreen(
	buttonTexture2D rl.Texture2D,
	display int,
	startTexture2D rl.Texture2D,
	gameTimer Timer,
) bool {
	button := Button{
		id:        -1,
		texture:   buttonTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 220, Height: 100},
		position: rl.Vector2{
			X: float32(rl.GetMonitorWidth(display))/2 - 220/2,
			Y: float32(rl.GetMonitorHeight(display))/2 - 220/2,
		},
		color:  rl.Purple,
		status: 0,
	}
	winTime := time.Now()
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		mousePoint := rl.GetMousePosition()
		if button.CheckInput(mousePoint) {
			return true
		}

		rl.DrawTextureRec(
			startTexture2D,
			rl.Rectangle{X: 0, Y: 0, Width: 1600, Height: 900},
			rl.Vector2{X: float32(rl.GetMonitorWidth(display))/2 - 800, Y: float32(rl.GetMonitorHeight(display))/2 - 450},
			rl.Gray,
		)
		rl.DrawText(
			fmt.Sprintf(
				"You've Won!",
			),
			int32(rl.GetMonitorWidth(display)/2-500),
			int32(rl.GetMonitorHeight(display)/2-300),
			100,
			rl.White,
		)
		printYourTime(gameTimer, winTime, true, display)
		button.Draw()
		rl.EndDrawing()
	}
	return false
}

func countdown(countdownSound rl.Sound, display int, stageName string) {
	beginTimer := Timer{}
	beginTimer.Init()
	rl.PlaySound(countdownSound)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		secondsLeft := 1 - time.Since(beginTimer.gameInitTime).Seconds()

		rl.DrawText(
			fmt.Sprintf(
				"Be Ready for %s",
				stageName,
			),
			int32(rl.GetMonitorWidth(display)/2-300),
			int32(rl.GetMonitorHeight(display)/2-300),
			100,
			rl.Black,
		)
		if secondsLeft <= 0 {
			break
		}
		rl.EndDrawing()
	}
}

func hasWonStage() bool {
	for _, obj := range gameObjects {
		if obj.IsEnemy() {
			return false
		}
	}
	return true
}

func isMouseOverEnemy(mousePosition rl.Vector2) bool {
	for key, obj := range gameObjects {
		if obj.IsEnemy() {
			hitbox := obj.Hitbox()
			if mousePosition.X >= hitbox.X && mousePosition.X <= hitbox.X+hitbox.Width &&
				mousePosition.Y >= hitbox.Y && mousePosition.Y <= hitbox.Y+hitbox.Height {
				fmt.Printf("shot checked!!!")
				delete(gameObjects, key)
				return true
			}
		}
	}
	return false
}

func DrawGameObjects() {
	for _, obj := range gameObjects {
		obj.Draw()
	}
}

type GameObject interface {
	Draw()
	GameObjectId() int
	IsEnemy() bool
	Hitbox() rl.Rectangle
}

type Timer struct {
	gameInitTime time.Time
	position     rl.Vector2
}

func (t *Timer) Init() {
	t.gameInitTime = time.Now()
}

func printYourTime(gameTimer Timer, queryTime time.Time, won bool, display int) {
	duration := queryTime.Sub(gameTimer.gameInitTime)
	if won {
		rl.DrawText(
			fmt.Sprintf(
				"Your Record: %.0f s",
				duration.Seconds(),
			),
			int32(rl.GetMonitorWidth(display)/2+150),
			int32(rl.GetMonitorHeight(display)/2-250),
			50,
			rl.White,
		)
	} else {
		rl.DrawText(
			fmt.Sprintf(
				"Your Time: %.0f s",
				duration.Seconds(),
			),
			int32(gameTimer.position.X),
			int32(gameTimer.position.Y),
			100,
			rl.Black,
		)
	}
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

func (b *Button) IsEnemy() bool {
	return false
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

func (b *Button) Hitbox() rl.Rectangle {
	return b.sourceRec
}

type Player struct {
	id            int
	texture       rl.Texture2D
	sourceRec     rl.Rectangle
	position      rl.Vector2
	color         rl.Color
	movementSpeed float32
}

func (p *Player) Draw() {
	rl.DrawTextureRec(
		p.texture,
		p.sourceRec,
		p.position,
		p.color,
	)
}

func (p *Player) GameObjectId() int {
	return p.id
}

func (p *Player) IsEnemy() bool {
	return false
}

func (p *Player) Hitbox() rl.Rectangle {
	return p.sourceRec
}

type Enemy struct {
	id            int
	texture       rl.Texture2D
	sourceRec     rl.Rectangle
	position      rl.Vector2
	color         rl.Color
	movementSpeed float32
}

func (e *Enemy) Draw() {
	rl.DrawTextureRec(
		e.texture,
		e.sourceRec,
		e.position,
		e.color,
	)
}

func (e *Enemy) Hitbox() rl.Rectangle {
	return rl.Rectangle{
		X:      e.position.X,
		Y:      e.position.Y,
		Width:  e.sourceRec.Width,
		Height: e.sourceRec.Height,
	}
}

func (e *Enemy) GameObjectId() int {
	return e.id
}

func (e *Enemy) IsEnemy() bool {
	return true
}

func raylibWindowMidPoint(elementWidth float32, elementHeight float32) (midX float32, midY float32) {
	display := rl.GetCurrentMonitor()
	return float32(rl.GetMonitorWidth(display))/2 - elementWidth/2, float32(rl.GetMonitorHeight(display))/2 - elementHeight/2
}
