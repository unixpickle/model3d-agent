package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	mesh, colorFunc := CreateModel()
	SaveRandomGridWhiteBG("{tmp_image_name}", mesh, 3, 3, 300, colorFunc.RenderColor)
}

func SaveRandomGridWhiteBG(
	path string, obj any, rows, cols, imgSize int, colorFunc render3d.ColorFunc,
) error {
	object := render3d.Objectify(obj, colorFunc)
	fullOutput := render3d.NewImage(cols*imgSize, rows*imgSize)

	min, max := object.Min(), object.Max()
	center := min.Mid(max)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Less aggressive randomness in Z direction
			direction := model3d.NewCoord3DRandNorm().Mul(model3d.XYZ(1.0, 1.0, 0.25)).Normalize()

			caster := &render3d.RayCaster{
				Camera: render3d.DirectionalCamera(object, direction, math.Pi/3.6),
				Lights: []*render3d.PointLight{
					{
						Origin: center.Add(direction.Scale(1000)),
						Color:  render3d.NewColor(1.0),
					},
				},
			}
			subImage := render3d.NewImage(imgSize*2, imgSize*2)
			subImage.SetAll(render3d.NewColor(1))
			caster.Render(subImage, object)
			fullOutput.CopyFrom(subImage.Downsample(2), j*imgSize, i*imgSize)
		}
	}

	return fullOutput.Save(path)
}
