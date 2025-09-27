package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

// Subtle, smooth cactus trunk.
func CactusMainTrunk(height, baseR, tipR float64) model3d.Solid {
	waveCount := 1.5 // Very subtle tall undulation
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-baseR*1.13, -baseR*1.13, -0.16*baseR),
		model3d.XYZ(baseR*1.13, baseR*1.13, height+tipR*1.1),
		func(c model3d.Coord3D) bool {
			blend := c.Z / height
			if c.Z < 0 {
				// Rounded base
				baseBulge := math.Sqrt(math.Max(0, 1-(c.Z/(0.14*baseR))*(c.Z/(0.14*baseR)))) * 0.08
				radius := baseR + baseBulge
				return c.XY().Norm() < radius
			} else if c.Z > height {
				return false
			}
			theta := math.Atan2(c.Y, c.X)
			// Gentle undulation only
			wave := 0.018*math.Sin(waveCount*theta + 0.9*blend)
			radius := baseR*(1-blend) + tipR*blend + wave
			return c.XY().Norm() < radius
		},
	)
}

// Bezier S-curve for arms (much smoother).
func CactusArmBezier(base, ctrl, tip model3d.Coord3D, r0, r1 float64) model3d.Solid {
	segments := 14
	var pieces model3d.JoinedSolid
	for i := 0; i < segments; i++ {
		t0 := float64(i) / float64(segments)
		t1 := float64(i+1) / float64(segments)
		bezier := func(t float64) model3d.Coord3D {
			// Quadratic Bezier
			return base.Scale((1-t)*(1-t)).Add(ctrl.Scale(2*(1-t)*t)).Add(tip.Scale(t*t))
		}
		p0 := bezier(t0)
		p1 := bezier(t1)
		// Slight but smooth radius taper and tiny waviness for realism.
		blend := (t0 + t1) / 2
		r := r0*(1-blend) + r1*blend + 0.008*math.Sin(8*t0)
		rNext := r0*(1-(blend+1.0/float64(segments))/2) + r1*((blend+1.0/float64(segments))/2)
		pieces = append(pieces, &model3d.Cylinder{
			P1:     p0,
			P2:     p1,
			Radius: 0.5 * (r + rNext),
		})
	}
	// Spherical tip
	pieces = append(pieces, &model3d.Sphere{Center: tip, Radius: r1 * 1.09})
	return pieces
}

func CreateModel() (*model3d.Mesh, toolbox3d.CoordColorFunc) {
	mainHeight := 2.0
	mainRadius := 0.32
	mainRadiusTop := 0.19

	// Two classic arms: start lower, curve upward and outward but not horizontally straight out.
	arm1Base := model3d.XYZ(
		mainRadius*0.71*math.Cos(math.Pi*0.63),
		mainRadius*0.71*math.Sin(math.Pi*0.63),
		0.43*mainHeight,
	)
	arm2Base := model3d.XYZ(
		mainRadius*0.70*math.Cos(-math.Pi*0.57),
		mainRadius*0.70*math.Sin(-math.Pi*0.57),
		0.57*mainHeight,
	)
	armLen := 1.07
	armRadius := mainRadius * 0.50
	armTipRadius := armRadius * 0.74

	// Arm control and tip: Outward then curve up with control point below tip for S.
	toUnit := func(theta float64) model3d.Coord3D {
		return model3d.XY(math.Cos(theta), math.Sin(theta))
	}
	// Arm 1: left/upward
	arm1Ctrl := arm1Base.Add(toUnit(math.Pi*0.63).Scale(armLen * 0.42)).Add(model3d.Z(armLen * 0.37))
	arm1Tip := arm1Base.Add(toUnit(math.Pi*0.54).Scale(armLen * 0.76)).Add(model3d.Z(armLen * 0.85))
	arm1 := CactusArmBezier(arm1Base, arm1Ctrl, arm1Tip, armRadius, armTipRadius)

	// Arm 2: right/upward
	arm2Ctrl := arm2Base.Add(toUnit(-math.Pi*0.57).Scale(armLen * 0.39)).Add(model3d.Z(armLen * 0.33))
	arm2Tip := arm2Base.Add(toUnit(-math.Pi*0.63).Scale(armLen * 0.62)).Add(model3d.Z(armLen * 0.92))
	arm2 := CactusArmBezier(arm2Base, arm2Ctrl, arm2Tip, armRadius, armTipRadius)

	trunk := CactusMainTrunk(mainHeight, mainRadius, mainRadiusTop)
	trunkTop := &model3d.Sphere{
		Center: model3d.XYZ(0, 0, mainHeight+mainRadiusTop*0.60),
		Radius: mainRadiusTop * 1.10,
	}

	// Flatter, wider dirt base.
	dirtRadius := mainRadius * 1.25
	dirtHeight := dirtRadius * 0.35
	dirt := model3d.CheckedFuncSolid(
		model3d.XYZ(-dirtRadius, -dirtRadius, -dirtHeight-0.02),
		model3d.XYZ(dirtRadius, dirtRadius, 0.06),
		func(c model3d.Coord3D) bool {
			// Ellipsoidal, clipped at the top.
			if c.Z < -dirtHeight-0.02 || c.Z > 0.05 {
				return false
			}
			norm2 := c.XY().Norm()
			norm2 = norm2 * norm2
			return (norm2/(dirtRadius*dirtRadius) + ((c.Z)/(dirtHeight*1.04))*((c.Z)/(dirtHeight*1.04))) < 1.0
		},
	)

	solid := model3d.JoinedSolid{
		trunk,
		trunkTop,
		arm1,
		arm2,
		dirt,
	}

	delta := 0.021
	mesh, points := model3d.DualContourInterior(solid, delta, true, false)

	// Natural cactus colors.
	bodyColor := render3d.NewColorRGB(0.18, 0.68, 0.18)
	bodyStripes := render3d.NewColorRGB(0.29, 0.82, 0.35)
	bodyTipColor := render3d.NewColorRGB(0.69, 0.97, 0.81)
	dirtColor := render3d.NewColorRGB(0.43, 0.31, 0.13)

	coordTree := model3d.NewCoordTree(points)

	return mesh, func(c model3d.Coord3D) render3d.Color {
		p := coordTree.NearestNeighbor(c)
		if p.Z < 0.08 {
			return dirtColor
		}
		// Trunk top blend
		if p.Dist(trunkTop.Center) < trunkTop.Radius+0.055 {
			dist := p.Dist(trunkTop.Center)
			f := 1 - math.Min(dist/(trunkTop.Radius+0.013), 1)
			r1, g1, b1 := render3d.RGB(bodyColor)
			r2, g2, b2 := render3d.RGB(bodyTipColor)
			return render3d.NewColorRGB(
				r1*(1-f)+r2*f,
				g1*(1-f)+g2*f,
				b1*(1-f)+b2*f,
			)
		}
		// Arm tips blend
		for _, tip := range []model3d.Coord3D{arm1Tip, arm2Tip} {
			if p.Dist(tip) < armTipRadius+0.045 {
				frac := 1 - math.Min(p.Dist(tip)/(armTipRadius+0.01), 1)
				r1, g1, b1 := render3d.RGB(bodyColor)
				r2, g2, b2 := render3d.RGB(bodyTipColor)
				return render3d.NewColorRGB(
					r1*(1-frac)+r2*frac,
					g1*(1-frac)+g2*frac,
					b1*(1-frac)+b2*frac,
				)
			}
		}
		// Subtle vertical striping for cactus body
		theta := math.Atan2(p.Y, p.X)
		stripeVal := 0.11 + 0.84*math.Pow(0.5+0.5*math.Cos(theta*6.2+2.1*math.Sin(p.Z*0.8)), 1.8)
		r1, g1, b1 := render3d.RGB(bodyColor)
		r2, g2, b2 := render3d.RGB(bodyStripes)
		return render3d.NewColorRGB(
			r1*(1-stripeVal)+r2*stripeVal,
			g1*(1-stripeVal)+g2*stripeVal,
			b1*(1-stripeVal)+b2*stripeVal,
		)
	}
}
