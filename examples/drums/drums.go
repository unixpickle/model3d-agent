package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func drumBody(radius, height float64) model3d.Solid {
	return &model3d.Cylinder{
		P1: model3d.Z(0),
		P2: model3d.Z(height),
		Radius: radius,
	}
}

func drumHead(radius, z float64) model3d.Solid {
	return &model3d.Cylinder{
		P1: model3d.Z(z-0.005),
		P2: model3d.Z(z+0.005),
		Radius: radius + 0.018,
	}
}

func drumLugRing(radius, z float64) model3d.Solid {
	return &model3d.Torus{
		Center:      model3d.Z(z),
		Axis:        model3d.Z(1),
		OuterRadius: radius + 0.014,
		InnerRadius: 0.0035,
	}
}

func CreateModel() (*model3d.Mesh, toolbox3d.CoordColorFunc) {
	// Drums are axis-aligned for simplicity.

	// KIT MEASUREMENTS (all realistic-ish, in meters)
	bassRadius := 0.22
	bassLength := 0.44

	snareRadius := 0.12
	snareHeight := 0.13

	tom1Radius := 0.10
	tom1Height := 0.11
	tom2Radius := 0.09
	tom2Height := 0.10
	floorTomRadius := 0.14
	floorTomHeight := 0.15

	// Core Zs
	bassZ := bassRadius + 0.03 // bass on short feet
	floorZ := 0.0

	// -- Drum arrangement (all centered at shell center) --
	bassCenter := model3d.XYZ(0, 0, bassZ)
	snareCenter := bassCenter.Add(model3d.XYZ(-0.27, 0.30, snareHeight/2 - bassRadius + bassZ - 0.01))
	tom1Center := bassCenter.Add(model3d.XYZ(-0.07, 0.11, bassRadius + 0.06 + tom1Height/2))
	tom2Center := bassCenter.Add(model3d.XYZ(0.13, 0.13, bassRadius + 0.03 + tom2Height/2))
	floorTomCenter := bassCenter.Add(model3d.XYZ(0.29, -0.18, floorTomHeight/2))

	// Solids by color
	var redDrums model3d.JoinedSolid
	var whiteHeads model3d.JoinedSolid
	var chromeParts model3d.JoinedSolid
	var goldParts model3d.JoinedSolid

	// ---- Drum shells (red) ----
	// Bass (horizontal cylinder in Y)
	redDrums = append(redDrums, &model3d.Cylinder{
		P1: bassCenter.Add(model3d.Y(-bassLength / 2)),
		P2: bassCenter.Add(model3d.Y(bassLength / 2)),
		Radius: bassRadius,
	})
	// Snare
	redDrums = append(redDrums, model3d.TranslateSolid(
		drumBody(snareRadius, snareHeight),
		snareCenter.Sub(model3d.Z(snareHeight/2)),
	))
	// Rack tom 1
	redDrums = append(redDrums, model3d.TranslateSolid(
		drumBody(tom1Radius, tom1Height),
		tom1Center.Sub(model3d.Z(tom1Height/2)),
	))
	// Rack tom 2
	redDrums = append(redDrums, model3d.TranslateSolid(
		drumBody(tom2Radius, tom2Height),
		tom2Center.Sub(model3d.Z(tom2Height/2)),
	))
	// Floor tom
	redDrums = append(redDrums, model3d.TranslateSolid(
		drumBody(floorTomRadius, floorTomHeight),
		floorTomCenter.Sub(model3d.Z(floorTomHeight/2)),
	))

	// ---- Drum heads (white) ----
	// Bass heads (front/back)
	whiteHeads = append(whiteHeads,
		&model3d.Cylinder{
			P1: bassCenter.Add(model3d.Y(-bassLength/2 - 0.005)),
			P2: bassCenter.Add(model3d.Y(-bassLength/2 + 0.005)),
			Radius: bassRadius + 0.017,
		},
		&model3d.Cylinder{
			P1: bassCenter.Add(model3d.Y(bassLength/2 - 0.005)),
			P2: bassCenter.Add(model3d.Y(bassLength/2 + 0.005)),
			Radius: bassRadius + 0.017,
		},
		// Snare heads
		model3d.TranslateSolid(drumHead(snareRadius, 0), snareCenter.Sub(model3d.Z(snareHeight/2))),
		model3d.TranslateSolid(drumHead(snareRadius, snareHeight), snareCenter.Sub(model3d.Z(snareHeight/2))),
		// Tom heads
		model3d.TranslateSolid(drumHead(tom1Radius, tom1Height), tom1Center.Sub(model3d.Z(tom1Height/2))),
		model3d.TranslateSolid(drumHead(tom2Radius, tom2Height), tom2Center.Sub(model3d.Z(tom2Height/2))),
		model3d.TranslateSolid(drumHead(floorTomRadius, floorTomHeight), floorTomCenter.Sub(model3d.Z(floorTomHeight/2))),
	)

	// ---- Chrome hardware ----
	// Bass drum hoops (torus at front/back)
	for _, y := range []float64{-bassLength/2, bassLength/2} {
		chromeParts = append(chromeParts, &model3d.Torus{
			Center:      bassCenter.Add(model3d.Y(y)),
			Axis:        model3d.Y(1),
			OuterRadius: bassRadius + 0.014,
			InnerRadius: 0.004,
		})
	}
	// Drum top rings
	chromeParts = append(chromeParts,
		model3d.TranslateSolid(drumLugRing(snareRadius, snareHeight), snareCenter.Sub(model3d.Z(snareHeight/2))),
		model3d.TranslateSolid(drumLugRing(tom1Radius, tom1Height), tom1Center.Sub(model3d.Z(tom1Height/2))),
		model3d.TranslateSolid(drumLugRing(tom2Radius, tom2Height), tom2Center.Sub(model3d.Z(tom2Height/2))),
		model3d.TranslateSolid(drumLugRing(floorTomRadius, floorTomHeight), floorTomCenter.Sub(model3d.Z(floorTomHeight/2))),
	)

	// Hardware legs -- snare and floor tom get 3 each, for realism
	for _, drum := range []struct {
		center model3d.Coord3D
		radius, height float64
		spread float64
		legZ float64
	}{
		{snareCenter, snareRadius, snareHeight, 0.8, floorZ},
		{floorTomCenter, floorTomRadius, floorTomHeight, 1.0, floorZ},
	} {
		for i := 0; i < 3; i++ {
			a := (2 * math.Pi / 3) * float64(i)
			offset := model3d.XY(math.Cos(a), math.Sin(a)).Scale(drum.radius * drum.spread)
			top := drum.center.Add(offset)
			bottomZ := drum.legZ
			bottom := model3d.XYZ(top.X, top.Y, bottomZ)
			chromeParts = append(chromeParts, &model3d.Cylinder{
				P1: top,
				P2: bottom,
				Radius: 0.009,
			})
		}
	}
	// Bass drum spurs
	for _, sign := range []float64{-1, 1} {
		side := bassCenter.Add(model3d.X(sign * bassRadius * 0.91)).Add(model3d.Y(-bassLength/2 * 0.82)).Add(model3d.Z(-bassRadius*0.98))
		// The foot touches z = floorZ
		foot := model3d.XYZ(side.X, side.Y-0.02, floorZ)
		chromeParts = append(chromeParts, &model3d.Cylinder{
			P1: side,
			P2: foot,
			Radius: 0.012,
		})
	}
	// ---- Stands ----
	// Define cymbal stand tops/end positions
	// Compact, all start at floorZ
	standData := []struct {
		tip model3d.Coord3D
		base model3d.Coord3D
	}{
		// Crash (left)
		{
			bassCenter.Add(model3d.XYZ(-0.42, 0.05, bassRadius + 0.36)),
			model3d.XYZ(-0.44, -0.04, floorZ),
		},
		// Ride (right)
		{
			bassCenter.Add(model3d.XYZ(0.47, -0.23, bassRadius + 0.34)),
			model3d.XYZ(0.57, -0.37, floorZ),
		},
		// Hihat (left front)
		{
			bassCenter.Add(model3d.XYZ(-0.32, 0.44, 0.26)),
			model3d.XYZ(-0.32, 0.435, floorZ),
		},
	}
	for _, s := range standData {
		chromeParts = append(chromeParts, &model3d.Cylinder{
			P1: s.base,
			P2: s.tip,
			Radius: 0.010,
		})
	}

	// ---- Cymbals (gold, thinner, smaller) ----
	cymbalThickness := 0.005
	crashCymbal := model3d.RotateSolid(
		&model3d.Cylinder{
			P1: model3d.Z(-cymbalThickness/2),
			P2: model3d.Z(cymbalThickness/2),
			Radius: 0.16,
		},
		model3d.X(1), 0.20,
	)
	crashCymbal = model3d.TranslateSolid(crashCymbal, bassCenter.Add(model3d.XYZ(-0.42, 0.05, bassRadius + 0.36)))
	goldParts = append(goldParts, crashCymbal)
	rideCymbal := model3d.RotateSolid(
		&model3d.Cylinder{
			P1: model3d.Z(-cymbalThickness/2),
			P2: model3d.Z(cymbalThickness/2),
			Radius: 0.17,
		},
		model3d.Y(1), -0.22,
	)
	rideCymbal = model3d.TranslateSolid(rideCymbal, bassCenter.Add(model3d.XYZ(0.47, -0.23, bassRadius + 0.34)))
	goldParts = append(goldParts, rideCymbal)
	// Hi-hat (two disks, slightly apart)
	hiHatTop := &model3d.Cylinder{
		P1: bassCenter.Add(model3d.XYZ(-0.32, 0.44, 0.264)),
		P2: bassCenter.Add(model3d.XYZ(-0.32, 0.44, 0.269)),
		Radius: 0.13,
	}
	hiHatBot := &model3d.Cylinder{
		P1: bassCenter.Add(model3d.XYZ(-0.32, 0.44, 0.253)),
		P2: bassCenter.Add(model3d.XYZ(-0.32, 0.44, 0.258)),
		Radius: 0.13,
	}
	goldParts = append(goldParts, hiHatTop, hiHatBot)

	// --- All parts, mesh ---
	var allParts model3d.JoinedSolid
	allParts = append(allParts, redDrums...)
	allParts = append(allParts, whiteHeads...)
	allParts = append(allParts, chromeParts...)
	allParts = append(allParts, goldParts...)

	mesh, points := model3d.DualContourInterior(allParts, 0.010, true, false)

	cf := toolbox3d.JoinedSolidCoordColorFunc(
		points,
		redDrums, func(model3d.Coord3D) render3d.Color { return render3d.NewColorRGB(0.92, 0.07, 0.10) },
		whiteHeads, func(model3d.Coord3D) render3d.Color { return render3d.NewColorRGB(0.98, 0.98, 0.98) },
		chromeParts, func(model3d.Coord3D) render3d.Color { return render3d.NewColorRGB(0.7, 0.7, 0.72) },
		goldParts, func(model3d.Coord3D) render3d.Color { return render3d.NewColorRGB(0.98, 0.87, 0.17) },
	)
	return mesh, cf
}
