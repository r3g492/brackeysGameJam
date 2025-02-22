package main

import (
	"embed"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//go:embed resources/*
var resFS embed.FS

var (
	gameObjects                = make(map[int]GameObject)
	deadObjects                = make(map[int]GameObject)
	nextGameObjectId int       = 0
	nextDeadObjectId           = 0
	lastShotFired    time.Time = time.Now()
	stageEnd                   = 15
	minDistance      float32   = 1000
)

func LoadTextureFromEmbedded(filename string, resizeWidth int32, resizeHeight int32) (rl.Texture2D, *rl.Image) {
	data, err := resFS.ReadFile("resources/" + filename)
	if err != nil {
		log.Fatalf("failed to read embedded file %s: %v", filename, err)
	}
	ext := filepath.Ext(filename)
	img := rl.LoadImageFromMemory(ext, data, int32(len(data)))
	if img.Width == 0 {
		log.Fatalf("failed to load image %s from embedded data", filename)
	}
	if resizeWidth != -1 && resizeHeight != -1 {
		rl.ImageResize(img, resizeWidth, resizeHeight)
	}
	tex := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	return tex, img
}

func LoadSoundFromEmbedded(filename string) rl.Sound {
	data, err := resFS.ReadFile("resources/" + filename)
	if err != nil {
		log.Fatalf("failed to read embedded sound %s: %v", filename, err)
	}
	tmpFile, err := os.CreateTemp("", "*.mp3")
	if err != nil {
		log.Fatalf("failed to create temporary file for %s: %v", filename, err)
	}
	_, err = tmpFile.Write(data)
	if err != nil {
		log.Fatalf("failed to write to temporary file for %s: %v", filename, err)
	}
	tmpFile.Close()
	if err != nil {
		log.Fatalf("failed to rename temporary file: %v", err)
	}
	snd := rl.LoadSound(tmpFile.Name())
	return snd
}

func main() {
	display := rl.GetCurrentMonitor()
	userMonitorWidth := rl.GetMonitorWidth(display)
	userMonitorHeight := rl.GetMonitorHeight(display)
	screenWidth := int32(userMonitorWidth)
	screenHeight := int32(userMonitorHeight)
	rl.InitWindow(screenWidth, screenHeight, "The Cold Killer")
	rl.MaximizeWindow()
	if !rl.IsWindowFullscreen() {
		rl.ToggleFullscreen()
	}
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	screenWidth = int32(rl.GetScreenWidth())
	screenHeight = int32(rl.GetScreenHeight())

	buttonTexture2D, _ := LoadTextureFromEmbedded("button.png", -1, -1)
	startTexture2D, _ := LoadTextureFromEmbedded("start.png", 1600, 900)
	simpleTexture, _ := LoadTextureFromEmbedded("diamond.png", -1, -1)
	enemyTexture, _ := LoadTextureFromEmbedded("enemy.png", 100, 100)
	backgroundTexture, _ := LoadTextureFromEmbedded("snow.png", screenWidth, screenHeight)

	playerBackTexture, _ := LoadTextureFromEmbedded("Hero_back.png", 100, 100)
	playerFrontTexture, _ := LoadTextureFromEmbedded("Hero_front.png", 100, 100)
	playerLeftTexture, _ := LoadTextureFromEmbedded("Hero_left.png", 100, 100)
	playerRightTexture, _ := LoadTextureFromEmbedded("Hero_right.png", 100, 100)
	if !startButtonScreen(buttonTexture2D, display, startTexture2D) {
		return
	}

	midPointX, midPointY := raylibWindowMidPoint(100, 100)
	player := Player{
		id:            0,
		texture:       simpleTexture,
		sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 30, Height: 30},
		position:      rl.Vector2{X: midPointX, Y: midPointY},
		color:         rl.Black,
		movementSpeed: 15,
		backTexture:   playerBackTexture,
		frontTexture:  playerFrontTexture,
		leftTexture:   playerLeftTexture,
		rightTexture:  playerRightTexture,
		movement:      0,
	}
	gameObjects[0] = &player
	nextGameObjectId = 1

	// https://pixabay.com/music/trap-spinning-head-271171/
	bgm := LoadSoundFromEmbedded("spinning-head-271171.mp3")
	// https://pixabay.com/sound-effects/you-lose-game-sound-230514/
	loseSound := LoadSoundFromEmbedded("you-lose-game-sound-230514.mp3")
	// https://pixabay.com/sound-effects/game-bonus-2-294436/
	winSound := LoadSoundFromEmbedded("game-bonus-2-294436.mp3")
	// https://pixabay.com/sound-effects/shotgun-03-38220/
	gunShot := LoadSoundFromEmbedded("shotgun-03-38220.mp3")
	// https://pixabay.com/sound-effects/female-vocal-321-countdown-240912/
	countdownSound := LoadSoundFromEmbedded("female-vocal-321-countdown-240912.mp3")

	rl.PlaySound(bgm)
	stageIdx := 0

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

		countdown(countdownSound, display, strconv.Itoa(stageIdx+1))

		player.position.X = midPointX
		player.position.Y = midPointY
		CleanAllDead()
		// create enemy
		for i := 0; i <= stageIdx; i++ {
			enemyPosition := generateEnemyPosition(
				rl.Vector2{
					X: midPointX,
					Y: midPointY,
				},
				100,
				100,
				minDistance,
			)
			createEnemy(enemyTexture, enemyPosition)
		}

		for !rl.WindowShouldClose() {
			if rl.WindowShouldClose() {
				return
			}

			if hasWonStage() {
				if stageIdx >= stageEnd-1 {
					rl.PlaySound(winSound)
					rl.StopSound(bgm)
					if WinScreen(buttonTexture2D, display, startTexture2D, gameTimer) {
						return
					}
				}
				break
			}

			playerMovement(&player)
			if playerDeathCheck(&player) {
				rl.PlaySound(loseSound)
				rl.StopSound(bgm)
				if gameOverScreen(buttonTexture2D, display, startTexture2D) {
					CleanAllDead()
					// restart game
					stageIdx = -1
					gameTimer.Init()
					CleanAllEnemyAndBullet()
					break
				}
				return
			}

			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				if time.Since(lastShotFired) > time.Duration(200)*time.Millisecond {
					lastShotFired = time.Now()
					rl.PlaySound(gunShot)
					createBullet(simpleTexture, rl.GetMousePosition(), player)
				}
			}
			bulletCollisionCheck(enemyTexture)
			rl.BeginDrawing()
			rl.ClearBackground(rl.DarkGray)
			rl.DrawTexture(
				backgroundTexture,
				0,
				0,
				rl.Color{
					R: 150,
					G: 150,
					B: 150,
					A: 255,
				},
			)
			EnemyPlan(player)
			MoveGameObjects()
			DrawDeadObjects()
			DrawGameObjects()
			printYourTime(gameTimer, time.Now(), false, display)
			rl.EndDrawing()
		}
	}
}

func createBullet(diamondTexture2D rl.Texture2D, mousePosition rl.Vector2, player Player) {
	dx := mousePosition.X - player.position.X
	dy := mousePosition.Y - player.position.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance != 0 {
		unitX := dx / distance
		unitY := dy / distance
		bulletSpeed := float32(100)
		bulletVector := rl.Vector2{
			X: unitX * bulletSpeed,
			Y: unitY * bulletSpeed,
		}

		bullet := Bullet{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 10, Height: 10},
			position:      player.position,
			color:         rl.Yellow,
			movementSpeed: bulletSpeed,
			vector:        bulletVector,
		}
		gameObjects[nextGameObjectId] = &bullet
		nextGameObjectId++
	}
}

func createEnemy(enemyTexture rl.Texture2D, generatePosition rl.Vector2) {
	enemy := Enemy{
		id:               nextGameObjectId,
		texture:          enemyTexture,
		sourceRec:        rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
		position:         generatePosition,
		color:            rl.White,
		movementSpeed:    0,
		lastPlanVector:   rl.Vector2{},
		plan:             0,
		movePlan:         0,
		lastPlanInitTime: time.Now(),
		lastPlanDuration: time.Duration(100) * time.Millisecond,
		planSet:          false,
	}
	gameObjects[nextGameObjectId] = &enemy
	nextGameObjectId++
}

func createDead(deadTexture rl.Texture2D, generatePosition rl.Vector2) {
	dead := Dead{
		id:        nextDeadObjectId,
		texture:   deadTexture,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
		position:  generatePosition,
		color: color.RGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 30,
		},
		movementSpeed:    0,
		lastPlanVector:   rl.Vector2{},
		plan:             0,
		movePlan:         0,
		lastPlanInitTime: time.Now(),
		lastPlanDuration: time.Duration(100) * time.Millisecond,
		planSet:          false,
	}
	deadObjects[nextDeadObjectId] = &dead
	nextDeadObjectId++
}

func generateEnemyPosition(playerCenter rl.Vector2, enemyWidth, enemyHeight, minDistance float32) rl.Vector2 {
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())
	var pos rl.Vector2

	for {
		pos.X = rand.Float32() * (screenWidth - enemyWidth)
		pos.Y = rand.Float32() * (screenHeight - enemyHeight)

		enemyCenter := rl.Vector2{
			X: pos.X + enemyWidth/2,
			Y: pos.Y + enemyHeight/2,
		}

		dx := enemyCenter.X - playerCenter.X
		dy := enemyCenter.Y - playerCenter.Y
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		if distance >= minDistance {
			break
		}
	}
	return pos
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
		player.movement = 2
	}
	if isLeftPressed {
		player.position.X = player.position.X - dividedMovementSpeed
		player.movement = 3
	}
	if isDownPressed {
		player.position.Y = player.position.Y + dividedMovementSpeed
		player.movement = 0
	}
	if isRightPressed {
		player.position.X = player.position.X + dividedMovementSpeed
		player.movement = 1
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

	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()
	if playerHitbox.X+playerHitbox.Width < 0 || playerHitbox.X > float32(screenWidth) ||
		playerHitbox.Y+playerHitbox.Height < 0 || playerHitbox.Y > float32(screenHeight) {
		return true
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
		rl.DrawText(
			fmt.Sprintf(
				"The Cold Killer",
			),
			int32(rl.GetMonitorWidth(display)/2-500),
			int32(rl.GetMonitorHeight(display)/2-400),
			80,
			rl.Black,
		)
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
				"you died.",
			),
			int32(rl.GetMonitorWidth(display)/2-600),
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
				"%s / %s",
				stageName,
				strconv.Itoa(stageEnd),
			),
			int32(rl.GetMonitorWidth(display)/2-150),
			int32(rl.GetMonitorHeight(display)/2-100),
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

func bulletCollisionCheck(deadTexture rl.Texture2D) {
	for bulletKey, bulletObj := range gameObjects {
		if bulletObj.IsBullet() {
			bulletHitbox := bulletObj.Hitbox()

			if bulletHitbox.X < 0 || bulletHitbox.Y < 0 ||
				bulletHitbox.X > 5000 || bulletHitbox.Y > 5000 {
				delete(gameObjects, bulletKey)
				continue
			}

			bulletCurPos := rl.Vector2{X: bulletHitbox.X, Y: bulletHitbox.Y}
			bulletPrevPos := bulletObj.PrevPosition()

			for enemyKey, enemyObj := range gameObjects {
				if bulletKey == enemyKey {
					continue
				}
				if enemyObj.IsEnemy() {
					enemyHitbox := enemyObj.Hitbox()
					if rl.CheckCollisionRecs(bulletHitbox, enemyHitbox) ||
						lineIntersectsRect(bulletPrevPos, bulletCurPos, enemyHitbox) {
						delete(gameObjects, bulletKey)
						delete(gameObjects, enemyKey)
						createDead(deadTexture, enemyObj.PrevPosition())
						break
					}
				}
			}
		}
	}
}

func pointInRect(p rl.Vector2, rect rl.Rectangle) bool {
	return p.X >= rect.X && p.X <= rect.X+rect.Width &&
		p.Y >= rect.Y && p.Y <= rect.Y+rect.Height
}

func lineSegmentsIntersect(p, p2, q, q2 rl.Vector2) bool {
	orientation := func(a, b, c rl.Vector2) float32 {
		return (b.Y-a.Y)*(c.X-a.X) - (b.X-a.X)*(c.Y-a.Y)
	}
	o1 := orientation(p, p2, q)
	o2 := orientation(p, p2, q2)
	o3 := orientation(q, q2, p)
	o4 := orientation(q, q2, p2)

	return o1*o2 < 0 && o3*o4 < 0
}

func lineIntersectsRect(p, q rl.Vector2, rect rl.Rectangle) bool {
	if pointInRect(p, rect) || pointInRect(q, rect) {
		return true
	}

	topLeft := rl.Vector2{X: rect.X, Y: rect.Y}
	topRight := rl.Vector2{X: rect.X + rect.Width, Y: rect.Y}
	bottomRight := rl.Vector2{X: rect.X + rect.Width, Y: rect.Y + rect.Height}
	bottomLeft := rl.Vector2{X: rect.X, Y: rect.Y + rect.Height}

	if lineSegmentsIntersect(p, q, topLeft, topRight) ||
		lineSegmentsIntersect(p, q, topRight, bottomRight) ||
		lineSegmentsIntersect(p, q, bottomRight, bottomLeft) ||
		lineSegmentsIntersect(p, q, bottomLeft, topLeft) {
		return true
	}

	return false
}

func MoveGameObjects() {
	for _, obj := range gameObjects {
		obj.Move()
	}
}

func DrawGameObjects() {
	for _, obj := range gameObjects {
		obj.Draw()
	}
}

func DrawDeadObjects() {
	for _, obj := range deadObjects {
		obj.Draw()
	}
}

func CleanAllDead() {
	for _, obj := range deadObjects {
		obj.Delete()
	}
}

func EnemyPlan(player Player) {
	for _, obj := range gameObjects {
		if obj.IsEnemy() {
			obj.EnemyPlan(player)
		}
	}
}

func CleanAllEnemyAndBullet() {
	for _, obj := range gameObjects {
		if obj.IsEnemy() || obj.IsBullet() {
			obj.Delete()
		}
	}
}

type GameObject interface {
	Draw()
	GameObjectId() int
	IsEnemy() bool
	IsBullet() bool
	Hitbox() rl.Rectangle
	Move()
	PrevPosition() rl.Vector2
	EnemyPlan(player Player)
	Delete()
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
	rl.DrawText(
		fmt.Sprintf(
			"go",
		),
		int32(b.position.X+80),
		int32(b.position.Y+25),
		40,
		rl.Color{
			R: 250,
			G: 200,
			B: 0,
			A: 200,
		},
	)
}

func (b *Button) IsEnemy() bool {
	return false
}

func (b *Button) IsBullet() bool {
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

func (b *Button) Move() {
	return
}

func (b *Button) EnemyPlan(player Player) {
	return
}

func (b *Button) PrevPosition() rl.Vector2 {
	return b.position
}

func (b *Button) Delete() {
	return
}

type Player struct {
	id            int
	texture       rl.Texture2D
	sourceRec     rl.Rectangle
	position      rl.Vector2
	color         rl.Color
	movementSpeed float32
	backTexture   rl.Texture2D
	frontTexture  rl.Texture2D
	leftTexture   rl.Texture2D
	rightTexture  rl.Texture2D
	// 0: front 1: right 2: back 3: left
	movement int
}

func (p *Player) Draw() {
	mousePosition := rl.GetMousePosition()
	playerPosition := p.position
	playerToMouseVector := rl.Vector2{
		X: mousePosition.X - playerPosition.X,
		Y: mousePosition.Y - playerPosition.Y,
	}

	// Calculate the angle (in degrees) of the vector from the player to the mouse
	angle := math.Atan2(float64(playerToMouseVector.Y), float64(playerToMouseVector.X)) * (180 / math.Pi)
	if angle < 0 {
		angle += 360
	}

	var texture rl.Texture2D

	// Define quadrants:
	// Bottom (frontTexture): 45° to 135°   [mouse is below the player]
	// Left   (leftTexture): 135° to 225°    [mouse is left of the player]
	// Top    (backTexture): 225° to 315°    [mouse is above the player]
	// Right  (rightTexture): otherwise      [mouse is right of the player]
	if angle >= 45 && angle < 135 {
		texture = p.frontTexture
	} else if angle >= 135 && angle < 225 {
		texture = p.leftTexture
	} else if angle >= 225 && angle < 315 {
		texture = p.backTexture
	} else {
		texture = p.rightTexture
	}

	// Calculate texture drawing position (adjust as needed)
	textureWidth := float32(texture.Width)
	textureHeight := float32(texture.Height)
	texturePosition := rl.Vector2{
		X: p.position.X - textureWidth/3,
		Y: p.position.Y - textureHeight/3,
	}

	rl.DrawTextureRec(
		texture,
		rl.Rectangle{
			Width:  textureWidth,
			Height: textureHeight,
		},
		texturePosition,
		rl.White, // white color means no tint
	)
}

func (p *Player) GameObjectId() int {
	return p.id
}

func (p *Player) IsEnemy() bool {
	return false
}

func (p *Player) IsBullet() bool {
	return false
}

func (p *Player) Hitbox() rl.Rectangle {
	return p.sourceRec
}

func (p *Player) Move() {
}

func (p *Player) EnemyPlan(player Player) {
}

func (p *Player) PrevPosition() rl.Vector2 {
	return p.position
}

func (p *Player) Delete() {
	return
}

type Enemy struct {
	id             int
	texture        rl.Texture2D
	sourceRec      rl.Rectangle
	position       rl.Vector2
	color          rl.Color
	movementSpeed  float32
	lastPlanVector rl.Vector2
	/**
	0: stop
	1: run to player
	2: move movePlan 0: up 1: up-right 2: right 3: right-down 4: down 5: down-left 6: left 7: left-up
	3: run to player with anger
	*/
	plan int
	/**
	movePlan 0: up 1: up-right 2: right 3: right-down 4: down 5: down-left 6: left 7: left-up
	*/
	movePlan         int
	lastPlanInitTime time.Time
	lastPlanDuration time.Duration
	planSet          bool
}

func (e *Enemy) resetPlan() {
	nextPlan := rand.Intn(4)
	if e.plan == nextPlan {
		nextPlan = rand.Intn(4)
		e.lastPlanDuration = time.Duration(1000) * time.Millisecond
	}
	e.plan = nextPlan

	if e.plan == 2 {
		e.movePlan = rand.Intn(8)
		e.lastPlanDuration = time.Duration(500) * time.Millisecond
	} else {
		e.movePlan = 0
		e.lastPlanDuration = time.Duration(50) * time.Millisecond
	}

	e.movementSpeed = float32(rand.Intn(15) + 5)
	if e.plan == 3 {
		e.movementSpeed += 5
		e.lastPlanDuration = time.Duration(500) * time.Millisecond
	}

	e.lastPlanInitTime = time.Now()
	e.planSet = false
}

func (e *Enemy) invokeRush() {
	e.movementSpeed = float32(rand.Intn(5) + 5)
	e.plan = 3
	e.movementSpeed += 25

	e.lastPlanInitTime = time.Now()
	e.lastPlanDuration = time.Duration(rand.Intn(3)+1) * time.Second
	e.planSet = false
}

func (e *Enemy) isPlanOver() bool {
	timeSince := time.Since(e.lastPlanInitTime)
	if timeSince > e.lastPlanDuration {
		return true
	}
	return false
}

func (e *Enemy) isOutOfMonitor() bool {
	hb := e.Hitbox()
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()
	if hb.X+hb.Width < 0 || hb.X > float32(screenWidth) ||
		hb.Y+hb.Height < 0 || hb.Y > float32(screenHeight) {
		return true
	}
	return false
}

func (e *Enemy) Draw() {
	if e.plan == 3 && e.movementSpeed >= 30 {
		rl.DrawTextureRec(
			e.texture,
			e.sourceRec,
			e.position,
			rl.Color{
				R: 255,
				G: 100,
				B: 100,
				A: 255,
			},
		)
	} else {
		rl.DrawTextureRec(
			e.texture,
			e.sourceRec,
			e.position,
			e.color,
		)
	}

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

func (e *Enemy) IsBullet() bool {
	return false
}

func (e *Enemy) Move() {
	e.position.X += e.lastPlanVector.X
	e.position.Y += e.lastPlanVector.Y
}

func (e *Enemy) EnemyPlan(player Player) {
	if e.isOutOfMonitor() {
		e.invokeRush()
	} else {
		if e.isPlanOver() {
			e.resetPlan()
		}
	}

	if e.planSet {
		return
	}
	switch e.plan {
	case 0:
		e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
	case 1:
		dx := player.position.X - e.position.X
		dy := player.position.Y - e.position.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist != 0 {
			unitX := dx / dist
			unitY := dy / dist
			e.lastPlanVector = rl.Vector2{
				X: unitX * e.movementSpeed,
				Y: unitY * e.movementSpeed,
			}
		} else {
			e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
		}
	case 2:
		var direction rl.Vector2
		switch e.movePlan {
		case 0:
			direction = rl.Vector2{X: 0, Y: -1}
		case 1:
			direction = rl.Vector2{X: 1, Y: -1}
		case 2:
			direction = rl.Vector2{X: 1, Y: 0}
		case 3:
			direction = rl.Vector2{X: 1, Y: 1}
		case 4:
			direction = rl.Vector2{X: 0, Y: 1}
		case 5:
			direction = rl.Vector2{X: -1, Y: 1}
		case 6:
			direction = rl.Vector2{X: -1, Y: 0}
		case 7:
			direction = rl.Vector2{X: -1, Y: -1}
		default:
			direction = rl.Vector2{X: 0, Y: 0}
		}

		mag := float32(math.Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y)))
		if mag != 0 {
			direction.X /= mag
			direction.Y /= mag
		}
		e.lastPlanVector = rl.Vector2{
			X: direction.X * e.movementSpeed,
			Y: direction.Y * e.movementSpeed,
		}
	case 3:
		dx := player.position.X - e.position.X
		dy := player.position.Y - e.position.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist != 0 {
			unitX := dx / dist
			unitY := dy / dist
			e.lastPlanVector = rl.Vector2{
				X: unitX * e.movementSpeed,
				Y: unitY * e.movementSpeed,
			}
		} else {
			e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
		}
	default:
		e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
	}
	e.planSet = true
}

func (e *Enemy) PrevPosition() rl.Vector2 {
	return e.position
}

func (e *Enemy) Delete() {
	delete(gameObjects, e.id)
	return
}

type Dead struct {
	id               int
	texture          rl.Texture2D
	sourceRec        rl.Rectangle
	position         rl.Vector2
	color            rl.Color
	movementSpeed    float32
	lastPlanVector   rl.Vector2
	plan             int
	movePlan         int
	lastPlanInitTime time.Time
	lastPlanDuration time.Duration
	planSet          bool
}

func (d *Dead) Draw() {
	rl.DrawTextureRec(
		d.texture,
		d.sourceRec,
		d.position,
		d.color,
	)
}

func (d *Dead) Hitbox() rl.Rectangle {
	return rl.Rectangle{
		X:      d.position.X,
		Y:      d.position.Y,
		Width:  d.sourceRec.Width,
		Height: d.sourceRec.Height,
	}
}

func (d *Dead) GameObjectId() int {
	return d.id
}

func (d *Dead) IsEnemy() bool {
	return false
}

func (d *Dead) IsBullet() bool {
	return false
}

func (d *Dead) Move() {
}

func (d *Dead) EnemyPlan(player Player) {
}

func (d *Dead) PrevPosition() rl.Vector2 {
	return d.position
}

func (d *Dead) Delete() {
	delete(deadObjects, d.id)
	return
}

type Bullet struct {
	id            int
	texture       rl.Texture2D
	sourceRec     rl.Rectangle
	position      rl.Vector2
	color         rl.Color
	movementSpeed float32
	vector        rl.Vector2
}

func (b *Bullet) Draw() {
	rl.DrawTextureRec(
		b.texture,
		b.sourceRec,
		b.position,
		b.color,
	)
}

func (b *Bullet) Hitbox() rl.Rectangle {
	return rl.Rectangle{
		X:      b.position.X,
		Y:      b.position.Y,
		Width:  b.sourceRec.Width,
		Height: b.sourceRec.Height,
	}
}

func (b *Bullet) GameObjectId() int {
	return b.id
}

func (b *Bullet) IsEnemy() bool {
	return false
}

func (b *Bullet) IsBullet() bool {
	return true
}

func (b *Bullet) Move() {
	b.position.X += b.vector.X
	b.position.Y += b.vector.Y
}

func (b *Bullet) EnemyPlan(player Player) {
}

func (b *Bullet) PrevPosition() rl.Vector2 {
	return rl.Vector2{
		X: b.position.X - b.vector.X,
		Y: b.position.Y - b.vector.Y,
	}
}

func (b *Bullet) Delete() {
	delete(gameObjects, b.id)
	return
}

func raylibWindowMidPoint(elementWidth float32, elementHeight float32) (midX float32, midY float32) {
	display := rl.GetCurrentMonitor()
	return float32(rl.GetMonitorWidth(display))/2 - elementWidth/2, float32(rl.GetMonitorHeight(display))/2 - elementHeight/2
}
