package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"log"
	"runtime"
	"strings"
	"time"
)

const (
	NUM_BYTES_IN_32_BIT = 4
	width              = 640
	height             = 480
	vertexShaderSource = `
    #version 410

    uniform float u_time;

    in vec3 vp;
    void main() {
    		float pct = abs(sin(u_time));
        gl_Position = vec4(vp, pct);
    }
` + "\x00"

	fragmentShaderSource = `
    #version 410

    uniform vec2 u_resolution;
    uniform float u_time;

    vec3 colorA = vec3(0.149,0.141,0.912);
    vec3 colorB = vec3(1.000,0.833,0.224);

    out vec4 FragColor;

    void main() {
        vec3 color = vec3(0.0);

        float pct = abs(sin(u_time));

        // Mix uses pct (a value from 0-1) to
        // mix the two colors
        color = mix(colorA, colorB, pct);

        FragColor = vec4(color,1.0);
    }
` + "\x00"
)

var (
	right = []float32 {
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
	}
	square = []float32 {
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,
		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}
	isoceles = []float32{
		0, 0.5, 0, // top
		-0.5, -0.5, 0, // left
		0.5, -0.5, 0, // right
	}
)

func init() {
	// "ensures we will always execute in the same operating system thread"
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	defer glfw.Terminate()

	if err = gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL Version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()

	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	start := time.Now()

	cells := makeCells()

	for !window.ShouldClose() {
		gl.Uniform1f(gl.GetUniformLocation(prog, gl.Str("u_time\x00")), float32(time.Since(start).Seconds()))
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.UseProgram(prog)

		for x := range cells {
			for _, c := range cells[x] {
				c.draw()
			}
		}



		glfw.PollEvents()
		window.SwapBuffers()
	}

}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32 // is this actually an address?
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, NUM_BYTES_IN_32_BIT*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao) // https://www.khronos.org/registry/OpenGL-Refpages/gl4/html/glGenVertexArrays.xhtml
	gl.BindVertexArray(vao) // https://www.khronos.org/registry/OpenGL-Refpages/gl4/html/glBindVertexArray.xhtml

	// As best I can tell, GenVertexArrays registers a vertex array object with a 'name'
	// which is the value of our vao variable -- printing out the value here gives 1.
	// BindVertexArray takes in an "array name" -- perhaps under the hood, GenVertexArrays is
	// creating some space in heap memory that we can't access directly, and giving us back an
	// "address" of 1, and BindVertexArray is 

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}


func (c *cell) draw() {
		gl.BindVertexArray(c.drawable)
		gl.DrawArrays(gl.LINE_LOOP, 0, int32(len(square)/3))
}
