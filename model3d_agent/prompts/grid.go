package main

import "github.com/unixpickle/model3d/render3d"

func main() {
	mesh, colorFunc := CreateModel()
	render3d.SaveRandomGrid("{tmp_image_name}", mesh, 3, 3, 300, colorFunc.RenderColor)
}
