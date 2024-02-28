package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"runtime"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var (
	width  int
	height int
)


var (
// 	vertexShaderSource = `
// #version 330
// in vec2 position;
// void main() {
//     gl_Position = vec4(position, 0.0, 1.0);
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
	#define ENABLE 0


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

	const vec3 palette[PALETTE_SIZE] = vec3[](
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

// 	shaderCode = `
// #version 330
// // ... (Your GLSL shader code)
// `

	// Specify the paths for your input images
	inputImagePath   = "input.png"
	matrixImagePath  = "img/matrix_128x128.png"
	outputImagePath  = "output.png"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(1, 1, "Offscreen Rendering", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	// Load input image
	inputImage, err := imaging.Open(inputImagePath)
	if err != nil {
		log.Fatal(err)
	}

	width = inputImage.Bounds().Dx()
	height = inputImage.Bounds().Dy()


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

	// Create an offscreen framebuffer
	var framebuffer uint32
	gl.GenFramebuffers(1, &framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)

	// Create a texture to render to
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

	// Check if framebuffer is complete
	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("Framebuffer error: %x", status)
	}

	// Main render loop
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Set up uniforms (iChannel0, iChannel1, etc.)
	// ...

	// Draw a quad (two triangles) to cover the whole framebuffer
	// ...

	// Read pixels from framebuffer
	pixels := make([]uint8, width*height*4)
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	// Create an image from pixel data
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Pix = pixels

	// Save the image to a file
	outputFile, err := os.Create(outputImagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, img); err != nil {
		log.Fatal(err)
	}
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
