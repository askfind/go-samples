// Modfied by plumbum (2015)
// It want go 1.5+ for compilation

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

// An app that draws a green triangle on a red background.
//
// Note: This demo is an early preview of Go 1.5. In order to build this
// program as an Android APK using the gomobile tool.
//
// See http://godoc.org/golang.org/x/mobile/cmd/gomobile to install gomobile.
//
// Get the basic example and use gomobile to build or install it on your device.
//
//   $ go get -d golang.org/x/mobile/example/basic
//   $ gomobile build golang.org/x/mobile/example/basic # will build an APK
//
//   # plug your Android device to your computer or start an Android emulator.
//   # if you have adb installed on your machine, use gomobile install to
//   # build and deploy the APK to an Android target.
//   $ gomobile install golang.org/x/mobile/example/basic
//
// Switch to your device or emulator to start the Basic application from
// the launcher.
// You can also run the application on your desktop by running the command
// below. (Note: It currently doesn't work on Windows.)
//   $ go install golang.org/x/mobile/example/basic && basic
package main

import (
	"encoding/binary"
	"log"
	"math"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/event/size"
)

var (
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer

	green     float32
	direction float32 = 0.01
	touchX    float32
	touchY    float32
)

func main() {
	app.Main(func(a app.App) {
		var c size.Event
		for e := range a.Events() {
			switch e := app.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					onStart()
				case lifecycle.CrossOff:
					onStop()
				}
			case size.Event:
				c = e
				touchX = float32(c.WidthPx / 2)
				touchY = float32(c.HeightPx / 2)
			case paint.Event:
				circleTriangle()
				onPaint(c)
				a.EndPaint(e)
			case touch.Event:
				touchX = e.X
				touchY = e.Y
			}
		}
	})
}

var angle float32 = 0

func getXY(angle float32) (x, y float32) {

	return float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))

}

func circleTriangle() {

	x1, y1 := getXY(angle)
	x2, y2 := getXY(angle + math.Pi*2*1/3)
	x3, y3 := getXY(angle + math.Pi*2*2/3)

	triangleData = f32.Bytes(binary.LittleEndian,
		x1*0.4, y1*0.4, 0.0,
		x2*0.4, y2*0.4, 0.0,
		x3*0.4, y3*0.4, 0.0,
	)

	angle += math.Pi / 200
	if angle > math.Pi*2 {
		angle -= math.Pi * 2
	}
}

func onStart() {
	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	position = gl.GetAttribLocation(program, "position")
	color = gl.GetUniformLocation(program, "color")
	offset = gl.GetUniformLocation(program, "offset")

	// TODO(crawshaw): the debug package needs to put GL state init here
	// Can this be an app.RegisterFilter call now??
}

func onStop() {
	gl.DeleteProgram(program)
	gl.DeleteBuffer(buf)
}

func onPaint(c size.Event) {
	gl.ClearColor(0, 0.3, 0.3, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)

	green += direction
	if green > 1 {
		green = 1
		direction = -direction
	}
	if green < 0.4 {
		green = 0.4
		direction = -direction
	}
	gl.Uniform4f(color, 0, green, 0, 1)

	gl.Uniform2f(offset, touchX/float32(c.WidthPx), touchY/float32(c.HeightPx))

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)

	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW) // ?

	gl.EnableVertexAttribArray(position)
	gl.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(position)

	debug.DrawFPS(c)
}

var triangleData = f32.Bytes(binary.LittleEndian,
	0.0, 0.4, 0.0, // top left
	0.0, -0.4, 0.0, // bottom left
	0.8, 0.0, 0.0, // bottom right
)

const (
	coordsPerVertex = 3
	vertexCount     = 3
)

const vertexShader = `#version 100
uniform vec2 offset;

attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
