package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width  = 800
	height = 600
)

var (
// 	vertexShaderSource = `
// #version 150 core
// in vec2 position;
// out vec2 TexCoord;
// void main() {
//     gl_Position = vec4(position, 0.0, 1.0);
//     TexCoord = (position + 1.0) / 2.0;
// }` + "\x00"

	fragmentShaderSource = `
	#version 330
	#define N 32                   // Number of iterations per fragment (higher N = more samples)
	#define RGB8(h) (vec3(h >> 16 & 0xFF, h >> 8 & 0xFF, h & 0xFF) / 255.0) 
	#define PALETTE_SIZE 4        // Number of colours in the palette
	#define ERROR_FACTOR 0.5       // Quantisation error coefficient (0 = no dithering)
	#define PIXEL_SIZE 1.0         // Size of pixels in the shader output
	#define ENABLE_SORT            // Choose whether to enable the sorting procedures
	// #define OPTIMISED_KNOLL        // Run an optimised version of the algorithm
	#define ENABLE 1


	#define INFINITY 3.4e38        // 'Infinity'


	#define RGB8(h) (vec3(h >> 16 & 0xFF, h >> 8 & 0xFF, h & 0xFF) / 255.0) 



	in vec2 TexCoord;
	out vec4 FragColor;


	
	// uniform lowp sampler2D source;
	uniform highp sampler2D iChannel0;
	uniform highp sampler2D iChannel1;
	varying highp vec2 qt_TexCoord0;
	varying highp vec2 qt_TexCoord1;
	uniform highp vec2 iResolution;

	uniform highp vec2 iChannelResolution;

	// varying highp float done;



	uniform highp vec2 iMouse;
	// uniform sampler2D iChannel0;

	// vec3 palette[PALETTE_SIZE];

#if ENABLE == 1

	const highp vec3 palette[PALETTE_SIZE] = highp vec3[](
		RGB8(0x1e1e2e), RGB8(0x313244), RGB8(0x45475a), RGB8(0xf5c2e7)
	);


	vec3 sRGBtoLinear(vec3 colour)
	{
		return colour * (colour * (colour * 0.305306011 + 0.682171111) + 0.012522878);
		// return colour;
	}

	// Get the luminance value of a given colour
	float getLuminance(vec3 colour) {
		return colour.r * 0.299 + colour.g * 0.587 + colour.b * 0.114;
	}

	int getClosestColour(vec3 inputColour) {
		float closestDistance = INFINITY;
		int closestColour = 0;
		// vec3 paletteColor;

		for (int i = 0; i < PALETTE_SIZE; i++) {
			// Assuming palette colors are predefined
			// vec3 paletteColor = palette[1];
			// if (i == 0) paletteColor = vec3(0.0); // Example color
			// else if (i == 1) paletteColor = vec3(1.0); // Example color
			// Define other palette colors similarly

			// Calculate the difference manually
			vec3 difference = inputColour - sRGBtoLinear(palette[i]);
			float distance = dot(difference, difference);

			if (distance < closestDistance) {
				closestDistance = distance;
				closestColour = i;
			}
		}

		return closestColour;
	}

	float sampleThreshold(vec2 coord) {
		// Sample the centre of the texel
		ivec2 pixel = ivec2(coord / PIXEL_SIZE) % ivec2(textureSize(iChannel1, 0).xy);
		vec2 uv = vec2(pixel) / vec2(textureSize(iChannel1, 0).xy);
		vec2 offset = 0.5 / vec2(textureSize(iChannel1, 0).xy);
		return texture2D(iChannel1, uv + offset).x * (N - 1);


	}




	void main() {



		highp vec3 colour = texture2D(iChannel0, qt_TexCoord0).rgb;

		// // Screen wipe effect
		// if (gl_FragCoord.x < iMouse.x) {
		// 	gl_FragColor = vec4(colour, 1.0);
		// 	return;
		// }

		// ====================================== //
		// Actual dithering algorithm starts here //
		// ====================================== //

		// Fill the candidate array
		int candidates[N];
		vec3 quantError = vec3(0.0);
		vec3 colourLinear = sRGBtoLinear(colour);

		for (int i = 0; i < N; i++) {
			vec3 goalColour = colourLinear + quantError * ERROR_FACTOR;
			int closestColour = getClosestColour(goalColour);

			candidates[i] = closestColour;
			quantError += colourLinear - sRGBtoLinear(palette[closestColour]);
		}

	#if defined(ENABLE_SORT)
		// Sort the candidate array by luminance (bubble sort)
		for (int i = N - 1; i > 0; i--) {
			for (int j = 0; j < i; j++) {
				if (getLuminance(palette[candidates[j]]) > getLuminance(palette[candidates[j+1]])) {
					// Swap the candidates
					int t = candidates[j];
					candidates[j] = candidates[j + 1];
					candidates[j + 1] = t;
				}
			}
		}
	#endif // ENABLE_SORT

		// Select from the candidate array, using the value in the threshold matrix
		int index = int(sampleThreshold(gl_FragCoord.xy));
		gl_FragColor = vec4(palette[candidates[index]], 1.0);


	}
			

	
#else
	void main() {
		vec4 sourceColor = texture2D(iChannel0, qt_TexCoord0);
		gl_FragColor = vec4(sourceColor);
	}
#endif` + "\x00"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)

	window, err := glfw.CreateWindow(width, height, "Go OpenGL Shader Example", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	// Compile shaders
	// vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatal(err)
	}

	// Link shaders into program
	program := gl.CreateProgram()
	// gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	gl.UseProgram(program)

	// Create a VAO and bind it
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Set up a framebuffer for offscreen rendering
	var framebuffer uint32
	gl.GenFramebuffers(1, &framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)



	// Create textures for iChannel0 and iChannel1
	iChannel0Texture, err := loadTexture("path/to/iChannel0/image.jpg")
	if err != nil {
		log.Fatal(err)
	}

	iChannel1Texture, err := loadTexture("path/to/iChannel1/image.jpg")
	if err != nil {
		log.Fatal(err)
	}



	// // Create a texture to render to
	// var texture uint32
	// gl.GenTextures(1, &texture)
	// gl.BindTexture(gl.TEXTURE_2D, texture)
	// gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, iChannel0Texture, 0)

	// Check for framebuffer completeness
	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("Framebuffer incomplete, status: 0x%x", status)
	}

	// Main loop
	for !window.ShouldClose() {
		// Render to offscreen framebuffer
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Set up uniforms and render your scene here using the shader program
		gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("iChannel0\x00")), 0)
		gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("iChannel1\x00")), 1)

		// Activate and bind textures
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, iChannel0Texture)

		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, iChannel1Texture)

		// Render your scene using the shader program

		// Render to the default framebuffer
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render the offscreen texture using a full-screen quad

		window.SwapBuffers()
		glfw.PollEvents()
	}

	saveTextureAsImage(iChannel0Texture, "output_image.png")
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csource, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csource, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		logMessage := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &logMessage[0])

		return 0, fmt.Errorf("failed to compile %v: %v", source, string(logMessage))
	}

	return shader, nil
}





func loadTexture(filename string) (uint32, error) {
	img, err := loadImage(filename)
	if err != nil {
		return 0, err
	}

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()), 0,
		gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return texture, nil
}

func loadImage(filename string) (*image.RGBA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return rgba, nil
}

func saveTextureAsImage(texture uint32, filename string) {
	width, height := getWindowSize()
	data := make([]uint8, width*height*4)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	img.Pix = data

	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, img)
	if err != nil {
		log.Fatal(err)
	}
}

func getWindowSize() (width, height int) {
	return glfw.GetCurrentContext().GetFramebufferSize()
}






