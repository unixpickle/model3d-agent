package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const fieldOfView = math.Pi / 3.6

func main() {
	var frames int
	var imgSize int
	var zoom float64
	flag.IntVar(&frames, "frames", 32, "number of frames")
	flag.IntVar(&imgSize, "size", 300, "size of images")
	flag.Float64Var(&zoom, "zoom", 1.3, "get closer by this fraction")
	flag.Parse()

	axis := model3d.Z(1)
	cameraDir := model3d.YZ(1, 0.3)

	mesh, colorFunc := CreateModel()
	object := render3d.Objectify(mesh, colorFunc.RenderColor)
	min, max := object.Min(), object.Max()
	center := min.Mid(max)

	var furthestCamera *render3d.Camera
	transformed := make([]render3d.Object, frames)
	for i := 0; i < frames; i++ {
		theta := math.Pi * 2 * float64(i) / float64(frames)
		xObj := render3d.Translate(object, center.Scale(-1))
		xObj = render3d.Rotate(xObj, axis, theta)
		xObj = render3d.Translate(xObj, center)
		transformed[i] = xObj

		camera := render3d.DirectionalCamera(transformed[i], cameraDir.Normalize(), fieldOfView)
		camera.Origin = center.Add(camera.Origin.Sub(center).Scale(1 / zoom))
		if furthestCamera == nil ||
			camera.Origin.Dist(center) > furthestCamera.Origin.Dist(center) {
			furthestCamera = camera
		}
	}

	offset := furthestCamera.Origin.Sub(center)
	lights := []*render3d.PointLight{
		{
			Origin: center.Add(offset.Scale(1000)),
			Color:  render3d.NewColor(1.0),
		},
	}

	for i, xObj := range transformed {
		caster := &render3d.RayCaster{
			Camera: furthestCamera,
			Lights: lights,
		}
		subImage := render3d.NewImage(imgSize, imgSize)
		subImage.SetAll(render3d.NewColor(1))
		caster.Render(subImage, xObj)
		subImage.Save(fmt.Sprintf("%03d.png", i))
	}
}
