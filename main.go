package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/veandco/go-sdl2/sdl"
     ."ast"
     ."gui"
)

var winWidth, winHeight = 1280, 720
var rows, cols, numPics int = 4, 4, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type guiState struct {
	zoom      bool
	zoomImage *sdl.Texture
	zoomTree  *picture
}

type rgba struct {
	r, g, b byte
}

type picture struct {
	r Node
	g Node
	b Node
}

func (p *picture) String() string {
	return "( Picture\n" + p.r.String() + "\n" + p.g.String() + "\n" + p.b.String() + " )"
}

func NewPicture() *picture {
	p := &picture{}

	p.r = GetRandomBaseNode()
	p.g = GetRandomBaseNode()
	p.b = GetRandomBaseNode()

	num := rand.Intn(15)
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomBaseNode())
	}

	num = rand.Intn(15)
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomBaseNode())
	}

	num = rand.Intn(15)
	for i := 0; i < num; i++ {
		p.b.AddRandom(GetRandomBaseNode())
	}

	for p.r.AddLeaf(GetRandomLeaf()) {

	}

	for p.g.AddLeaf(GetRandomLeaf()) {

	}

	for p.b.AddLeaf(GetRandomLeaf()) {

	}

	return p

}

func (p *picture) pickRandomColor() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return p.r
	case 1:
		return p.g
	case 2:
		return p.b
	default:
		panic("pickRandomColor failed")
	}
}

func cross(a *picture, b *picture) *picture {
	aCopy := &picture{CopyTree(a.r, nil), CopyTree(a.g, nil), CopyTree(a.b, nil)}
	aColor := aCopy.pickRandomColor()
	bColor := b.pickRandomColor()

	aIndex := rand.Intn(aColor.NodeCount())
	aNode, _ := GetNthNode(aColor, aIndex, 0)

	bIndex := rand.Intn(bColor.NodeCount())
	bNode, _ := GetNthNode(bColor, bIndex, 0)
	bNodeCopy := CopyTree(bNode, bNode.GetParent())

	ReplaceNode(aNode, bNodeCopy)
	return aCopy
}

func evolve(survivors []*picture) []*picture {
	newPics := make([]*picture, numPics)
	i := 0
	for i < len(survivors) {
		a := survivors[i]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)
		i++
	}
	for i < len(newPics) {
		a := survivors[rand.Intn(len(survivors))]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)

		i++
	}

	for _, pic := range newPics {
		r := rand.Intn(8)
		for i := 0; i < r; i++ {
			pic.Mutate()
		}
	}
	return newPics
}

func (p *picture) Mutate() {
	r := rand.Intn(3)
	var nodeToMutate Node

	switch r {
	case 0:
		nodeToMutate = p.r
	case 1:
		nodeToMutate = p.g
	case 2:
		nodeToMutate = p.b
	}

	count := nodeToMutate.NodeCount()
	r = rand.Intn(count)

	nodeToMutate, count = GetNthNode(nodeToMutate, r, 0)
	mutation := Mutate(nodeToMutate)

	if nodeToMutate == p.r {
		p.r = mutation
	} else if nodeToMutate == p.g {
		p.g = mutation
	} else if nodeToMutate == p.b {
		p.b = mutation
	}
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index] = c.g
		pixels[index] = c.b
	}
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func ASTToPixels(pic *picture, w, h int) []byte {
	scale := float32(255 / 2)
	offset := float32(-1.0 * scale)
	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	for yi := 0; yi < h; yi++ {

		y := float32(yi)/float32(h)*2 - 1

		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1

			r := pic.r.Eval(x, y)
			g := pic.g.Eval(x, y)
			b := pic.b.Eval(x, y)

			pixels[pixelIndex] = byte(r*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(g*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(b*scale - offset)
			pixelIndex++
			pixelIndex++
		}
	}
	return pixels
}

func saveTree(p *picture) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}

	biggestNumber := 0
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".apt") {
			numberStr := strings.TrimSuffix(name, ".apt")
			num, err := strconv.Atoi(numberStr)
			if err == nil {
				if num > biggestNumber {
					biggestNumber = num
				}
			}
		}
	}
	saveName := strconv.Itoa(biggestNumber+1) + ".apt"
	file, err := os.Create(saveName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprint(file, p.String())
}

func main() {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Evim", 200, 200, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32
	keyboardState := sdl.GetKeyboardState()
	prevKeyBoardState := make([]uint8, len(keyboardState))

	rand.Seed(time.Now().UTC().UnixNano())

	picTrees := make([]*picture, numPics)
	for i := range picTrees {
		picTrees[i] = NewPicture()
	}

	picWidth := int(float32(winWidth/cols) * float32(.9))
	picHeight := int(float32(winHeight/rows) * float32(.8))

	pixelsChannel := make(chan pixelResult, numPics)
	buttons := make([]*ImageButton, numPics)
	for i := range picTrees {
		go func(i int) {
			pixels := ASTToPixels(picTrees[i], picWidth*2, picHeight*2)
			pixelsChannel <- pixelResult{pixels, i}
		}(i)

	}

	evolveButtonTex := GetSinglePixelTex(renderer, sdl.Color{255, 255, 255, 0})
	evolveRect := sdl.Rect{int32(float32(winWidth)/2 - float32(picWidth)/2), int32(float32(winHeight) - float32(winHeight)*.10), int32(picWidth), int32(float32(picHeight) * .40)}
	evolveButton := NewImageButton(renderer, evolveButtonTex, evolveRect, sdl.Color{255, 255, 255, 0})


	mouseState := GetMouseState()
	state := guiState{false, nil, nil}
	args := os.Args
	if len(args) > 1 {
		fileBytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
		}

		fileStr := string(fileBytes)
		pictureNode := BeginLexing(fileStr)
		p := &picture{pictureNode.GetChildren()[0], pictureNode.GetChildren()[1], pictureNode.GetChildren()[2]}
		pixels := ASTToPixels(p, winWidth, winHeight)
		tex := pixelsToTexture(renderer, pixels, winWidth, winHeight)
		state.zoom = true
		state.zoomImage = tex
		state.zoomTree = p

	}

	for {
		frameStart := time.Now()

		mouseState.Update()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					touchX := int(e.X * float32(winWidth))
					touchY := int(e.Y * float32(winHeight))
					mouseState.X = touchX
					mouseState.Y = touchY
					mouseState.LeftButton = true
				}
			}
		}

		if !state.zoom {
			select {
			case pixelsAndIndex, ok := <-pixelsChannel:
				if ok {
					tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth*2, picHeight*2)
					xi := pixelsAndIndex.index % cols
					yi := (pixelsAndIndex.index - xi) / cols
					x := int32(xi * picWidth)
					y := int32(yi * picHeight)
					xPad := int32(float32(winWidth) * .1 / float32(cols+1))
					yPad := int32(float32(winHeight) * .1 / float32(rows+1))
					x += xPad * (int32(xi) + 1)
					y += yPad * (int32(yi) + 1)
					rect := &sdl.Rect{x, y, int32(picWidth), int32(picHeight)}
					button := NewImageButton(renderer, tex, *rect, sdl.Color{255, 255, 255, 0})
					buttons[pixelsAndIndex.index] = button
				}
			default:
			}

			renderer.Clear()
			for i, button := range buttons {
				if button != nil {
					button.Update(&mouseState)
					if button.WasLeftClicked {
						button.IsSelected = !button.IsSelected
					} else if button.WasRightClicked {
						zoomPixels := ASTToPixels(picTrees[i], winWidth*2, winHeight*2)
						zoomTex := pixelsToTexture(renderer, zoomPixels, winWidth*2, winHeight*2)
						state.zoomImage = zoomTex
						state.zoomTree = picTrees[i]
						state.zoom = true
					}
					button.Draw(renderer)
				}
			}
			evolveButton.Update(&mouseState)
			if evolveButton.WasLeftClicked {
				selectedPictures := make([]*picture, 0)
				for i, button := range buttons {
					if button.IsSelected {
						selectedPictures = append(selectedPictures, picTrees[i])
					}
				}
				if len(selectedPictures) != 0 {
					for i := range buttons {
						buttons[i] = nil
					}
					picTrees = evolve(selectedPictures)
					for i := range picTrees {
						go func(i int) {
							pixels := ASTToPixels(picTrees[i], picWidth*2, picHeight*2)
							pixelsChannel <- pixelResult{pixels, i}
						}(i)
					}
				}
			}
			evolveButton.Draw(renderer)
		} else {
			if !mouseState.RightButton && mouseState.PrevRightButton {
				state.zoom = false
			}
			if keyboardState[sdl.SCANCODE_S] == 0 && prevKeyBoardState[sdl.SCANCODE_S] != 0 {
				saveTree(state.zoomTree)
			}
			renderer.Copy(state.zoomImage, nil, nil)
		}
		renderer.Present()
		if keyboardState[sdl.SCANCODE_ESCAPE] != 0 {
			return
		}

		for i, v := range keyboardState {
			prevKeyBoardState[i] = v
		}

		elapsedTime = float32(time.Since(frameStart).Seconds())
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}

	}

}
