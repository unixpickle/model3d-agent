package main

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func sandBump(x, y, z float64) float64 {
	return 0.016*(
		math.Sin(12*x+8*z)+
			math.Sin(7*x+13*y+7*z)+
			math.Cos(9*y+4.5*z+6.1*x)) +
		0.012*math.Sin(25*x+2.2*y+7.5*z)*
			math.Cos(13.2*y+8.9*z+15*x)
}

func ShellImprint(center model3d.Coord3D, radius, depth float64, raised bool) model3d.Solid {
	nPetals := 6
	var shell model3d.JoinedSolid
	for i := 0; i < nPetals; i++ {
		th := float64(i)*2*math.Pi/float64(nPetals) + 0.11*float64(i)
		dir := model3d.XYZ(math.Cos(th)*0.7, math.Sin(th), 0)
		shell = append(shell, &model3d.Sphere{
			Center: center.Add(dir.Scale(radius * 0.62)),
			Radius: radius * (0.48 + 0.15*math.Sin(th*2)),
		})
	}
	// Add a central dome
	shell = append(shell, &model3d.Sphere{
		Center: center,
		Radius: radius * 0.60,
	})
	ret := model3d.JoinedSolid(shell)
	if !raised {
		return ret
	}
	return ret
}

func FancySandBase(baseRadius, thickness float64) model3d.Solid {
	rand.Seed(101)
	numBumps := 12
	edgeVars := make([]float64, numBumps)
	for i := range edgeVars {
		edgeVars[i] = 0.96 + (rand.Float64() * 0.08)
	}
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-baseRadius, -baseRadius, 0),
		model3d.XYZ(baseRadius, baseRadius, thickness*1.8),
		func(c model3d.Coord3D) bool {
			r := math.Hypot(c.X, c.Y)
			th := math.Atan2(c.Y, c.X)
			thI := float64(numBumps) * (math.Mod(th+2*math.Pi, 2*math.Pi)) / (2*math.Pi)
			i := int(thI) % numBumps
			frac := thI - math.Floor(thI)
			rim := edgeVars[i]*(1-frac) + edgeVars[(i+1)%numBumps]*frac
			limit := baseRadius * rim
			h := thickness * (1.0 - 0.32*(r/limit) + 0.18*math.Cos(th*5.2))
			return c.Z >= 0 && c.Z <= h && r <= limit
		},
	)
}

func FancyTower(center model3d.Coord3D, radius, height float64, shape string) model3d.Solid {
	switch shape {
	case "octagon":
		N := 8
		return model3d.CheckedFuncSolid(
			center.Add(model3d.XYZ(-radius, -radius, 0)),
			center.Add(model3d.XYZ(radius, radius, height+radius)),
			func(c model3d.Coord3D) bool {
				cc := c.Sub(center)
				if cc.Z < 0 || cc.Z > height {
					return false
				}
				rOnPoly := radius / math.Cos(math.Pi/float64(N))
				for i := 0; i < N; i++ {
					angle := math.Pi*2*float64(i)/float64(N) - math.Pi/float64(N)
					nml := model3d.XYZ(math.Cos(angle), math.Sin(angle), 0)
					if cc.X*nml.X+cc.Y*nml.Y > rOnPoly {
						return false
					}
				}
				return true
			},
		)
	default:
		return &model3d.Cylinder{
			P1: center,
			P2: center.Add(model3d.Z(height)),
			Radius: radius,
		}
	}
}

func FancyBridge(center model3d.Coord3D, length, width, height float64, archCount int) model3d.Solid {
	bridgeBody := model3d.NewRect(
		model3d.XYZ(center.X-length/2, center.Y-width/2, center.Z),
		model3d.XYZ(center.X+length/2, center.Y+width/2, center.Z+height),
	)
	var arches model3d.JoinedSolid
	for i := 0; i < archCount; i++ {
		x := center.X - length/2 + (float64(i)+0.5)*length/float64(archCount)
		arch := &model3d.Cylinder{
			P1:     model3d.XYZ(x, center.Y, center.Z-0.01),
			P2:     model3d.XYZ(x, center.Y, center.Z+height*0.40),
			Radius: width * 0.31,
		}
		arches = append(arches, arch)
	}
	return model3d.Subtract(bridgeBody, arches)
}

func DripSpire(base model3d.Coord3D, height, bottomRadius, taper float64) model3d.Solid {
	return model3d.CheckedFuncSolid(
		base.Add(model3d.XYZ(-bottomRadius, -bottomRadius, 0)),
		base.Add(model3d.XYZ(bottomRadius, bottomRadius, height)),
		func(p model3d.Coord3D) bool {
			c := p.Sub(base)
			h := c.Z
			if h < 0 || h > height {
				return false
			}
			rad := (1.0-h/height)*bottomRadius + taper*(h/height)
			bump := 0.05 * math.Sin(20*c.X + 8*h) * math.Sin(18*c.Y + 16*h)
			r := math.Hypot(c.X, c.Y)
			return r <= rad + bump*math.Pow(1.0-h/height,2)
		},
	)
}

func WavyFlag(poleBase model3d.Coord3D, poleHeight, width float64, colorIdx int, phase float64) model3d.Solid {
	staff := &model3d.Cylinder{
		P1: poleBase,
		P2: poleBase.Add(model3d.Z(poleHeight)),
		Radius: width * 0.13,
	}
	flagRect := model3d.CheckedFuncSolid(
		model3d.XYZ(-width/2, 0, 0),
		model3d.XYZ(width/2, width/6, width*0.65),
		func(c model3d.Coord3D) bool {
			wave := 0.12 * math.Sin(c.Z*5 + phase)
			return c.Y >= 0 && c.Y < width/6 &&
				c.X >= -width/2 && c.X <= width/2 &&
				c.Z >= 0 && c.Z <= width*0.62+wave
		},
	)
	flagShape := model3d.RotateSolid(flagRect, model3d.Z(1), math.Pi*0.07)
	flagShape = model3d.TranslateSolid(flagShape, poleBase.Add(model3d.Z(poleHeight-width/3)))
	return model3d.JoinedSolid{staff, flagShape}
}

func CreateModel() (*model3d.Mesh, toolbox3d.CoordColorFunc) {
	rand.Seed(2342)
	baseRadius := 1.10
	baseThick := 0.21

	mainBase := FancySandBase(baseRadius, baseThick)
	var sandLumps model3d.JoinedSolid
	numLumps := 9
	for i := 0; i < numLumps; i++ {
		th := rand.Float64() * 2 * math.Pi
		r := 0.78 * baseRadius * (0.8 + 0.19*rand.Float64())
		z := 0.08 * (0.45 + 0.55*rand.Float64())
		lump := &model3d.Sphere{
			Center: model3d.XYZ(math.Cos(th)*r, math.Sin(th)*r, z),
			Radius: 0.13 + rand.Float64()*0.08,
		}
		sandLumps = append(sandLumps, lump)
	}
	base := model3d.JoinedSolid{mainBase}
	base = append(base, sandLumps...)

	moatWidth := 0.13
	moatDepth := 0.09
	moat := model3d.CheckedFuncSolid(
		model3d.XYZ(-baseRadius-0.28, -baseRadius-0.28, -0.04),
		model3d.XYZ(baseRadius+0.28, baseRadius+0.28, 0.13),
		func(c model3d.Coord3D) bool {
			r := math.Hypot(c.X, c.Y)
			th := math.Atan2(c.Y, c.X)
			bump := 0.025*math.Cos(7*th)*math.Sin(r*3)
			limit := baseRadius + moatWidth + bump
			inner := baseRadius - moatWidth*0.6 + 0.013*math.Sin(4*th)
			maxZ := moatDepth + 0.03*math.Sin(8*th)
			if r > inner && r < limit && c.Z >= -0.02 && c.Z < maxZ {
				return true
			}
			return false
		},
	)
	moat = model3d.JoinedSolid{moat, model3d.NewRect(
		model3d.XYZ(-0.19, baseRadius-0.05, -0.02),
		model3d.XYZ(0.19, baseRadius+0.31, -0.02+0.11),
	)}
	baseMinusMoat := model3d.Subtract(base, moat)

	towerSpecs := []struct {
		theta, dist, radius, height float64
		shape                       string
		octagon                     bool
	}{
		{0,             baseRadius-0.28, 0.19, 0.47, "octagon", true},
		{math.Pi/3,     baseRadius-0.25, 0.15, 0.40, "circular", false},
		{2*math.Pi/3,   baseRadius-0.29, 0.20, 0.48, "octagon", true},
		{math.Pi,       baseRadius-0.23, 0.17, 0.38, "circular", false},
		{4*math.Pi/3,   baseRadius-0.25, 0.15, 0.41, "circular", false},
		{5*math.Pi/3,   baseRadius-0.29, 0.20, 0.46, "octagon", true},
	}
	var towers model3d.JoinedSolid
	for _, spec := range towerSpecs {
		pos := model3d.XYZ(spec.dist*math.Cos(spec.theta), spec.dist*math.Sin(spec.theta), baseThick)
		tw := FancyTower(pos, spec.radius, spec.height, spec.shape)
		var top model3d.Solid
		if spec.octagon {
			top = DripSpire(pos.Add(model3d.Z(spec.height)), 0.13+rand.Float64()*0.13, spec.radius*0.54, 0.01+rand.Float64()*0.04)
		} else {
			top = &model3d.Sphere{
				Center: pos.Add(model3d.Z(spec.height+spec.radius*0.53)),
				Radius: spec.radius * (0.57 + rand.Float64()*0.13),
			}
		}
		towers = append(towers, tw, top)
	}
	keepBase := model3d.XYZ(0, 0, baseThick)
	keepRadius := 0.31
	keepHeight := 0.63
	keep := FancyTower(keepBase, keepRadius, keepHeight, "octagon")
	keepSpire := &model3d.Cone{
		Base: keepBase.Add(model3d.Z(keepHeight)),
		Tip:  keepBase.Add(model3d.Z(keepHeight + 0.33)),
		Radius: keepRadius * 0.68,
	}
	towers = append(towers, keep, keepSpire)

	wallHeight := baseThick + 0.24
	wallThick := 0.13
	var walls model3d.JoinedSolid
	for i := 0; i < len(towerSpecs); i++ {
		a, b := towerSpecs[i], towerSpecs[(i+1)%len(towerSpecs)]
		ap := model3d.XYZ(a.dist*math.Cos(a.theta), a.dist*math.Sin(a.theta), baseThick)
		bp := model3d.XYZ(b.dist*math.Cos(b.theta), b.dist*math.Sin(b.theta), baseThick)
		mid := ap.Mid(bp)
		dir := bp.Sub(ap)
		lenWall := dir.Norm()
		theta := math.Atan2(dir.Y, dir.X)
		rect := model3d.NewRect(
			model3d.XYZ(-lenWall/2, -wallThick/2, baseThick),
			model3d.XYZ(lenWall/2, wallThick/2, wallHeight),
		)
		rotWall := model3d.RotateSolid(rect, model3d.Z(1), theta)
		transWall := model3d.TranslateSolid(rotWall, mid)
		walls = append(walls, transWall)
	}
	numSpurs := 3
	for i := 0; i < numSpurs; i++ {
		a := towerSpecs[i]
		ap := model3d.XYZ(a.dist*math.Cos(a.theta), a.dist*math.Sin(a.theta), baseThick)
		mid := ap.Mid(keepBase)
		dir := keepBase.Sub(ap)
		lenWall := dir.Norm()
		theta := math.Atan2(dir.Y, dir.X)
		rect := model3d.NewRect(
			model3d.XYZ(-lenWall/2, -wallThick*0.86/2, baseThick+0.01),
			model3d.XYZ(lenWall/2, wallThick*0.86/2, wallHeight-0.04),
		)
		rotWall := model3d.RotateSolid(rect, model3d.Z(1), theta)
		transWall := model3d.TranslateSolid(rotWall, mid)
		walls = append(walls, transWall)
	}

	var battlements model3d.JoinedSolid
	battlementSpecs := []struct{
		center model3d.Coord3D
		radius float64
		z      float64
		count  int
		width  float64
	}{
		{keepBase, keepRadius + 0.01, keepHeight + 0.03, 14, 0.11},
	}
	for _, spec := range towerSpecs {
		cp := model3d.XYZ(
			spec.dist*math.Cos(spec.theta),
			spec.dist*math.Sin(spec.theta),
			baseThick,
		)
		battlementSpecs = append(battlementSpecs, struct{
			center model3d.Coord3D
			radius float64
			z      float64
			count  int
			width  float64
		}{cp, spec.radius + 0.01, spec.height + baseThick + 0.01, 12, 0.075 + 0.021*rand.Float64()})
	}

	for _, b := range battlementSpecs {
		for j := 0; j < b.count; j++ {
			th := 2 * math.Pi * float64(j) / float64(b.count)
			height := 0.10 + 0.012*math.Sin(2*th)
			x, y := b.radius*math.Cos(th), b.radius*math.Sin(th)
			z := b.z
			merlon := model3d.NewRect(
				model3d.XYZ(-b.width/2, -0.062/2, 0),
				model3d.XYZ(b.width/2, 0.062/2, height),
			)
			rotMerlon := model3d.RotateSolid(merlon, model3d.Z(1), th)
			transMerlon := model3d.TranslateSolid(rotMerlon, b.center.Add(model3d.XYZ(x, y, z)))
			battlements = append(battlements, transMerlon)
			if j%5 == 0 && rand.Float64() < 0.67 {
				shell := ShellImprint(
					b.center.Add(model3d.XYZ(x, y, z + height*0.68)),
					b.width*0.31, 0.012, true)
				battlements = append(battlements, shell)
			}
		}
	}

	bridgeCenter := model3d.XYZ(0, baseRadius+moatWidth*0.5, 0.011)
	bridge := FancyBridge(bridgeCenter, 0.46, 0.12, 0.09, 3)

	stairBase := model3d.XYZ(0, baseRadius-0.08, 0)
	numStairs := 6
	stairWidth := 0.24
	stairDepth := 0.073
	stairHeight := (baseThick*0.95) / float64(numStairs)
	var stairs model3d.JoinedSolid
	for k := 0; k < numStairs; k++ {
		h0 := float64(k) * stairHeight
		rect := model3d.NewRect(
			model3d.XYZ(-stairWidth/2, stairBase.Y+float64(k)*stairDepth, h0),
			model3d.XYZ(stairWidth/2, stairBase.Y+float64(k+1)*stairDepth, h0+stairHeight+0.008),
		)
		stairs = append(stairs, rect)
	}
	for n := 0; n < 2; n++ {
		wth := 0.12 + 0.07*rand.Float64()
		depth := 0.06 + 0.03*rand.Float64()
		stH := 0.07
		xc := baseRadius*math.Cos(-math.Pi/3+float64(n+1)*0.73)
		yc := baseRadius*math.Sin(-math.Pi/3+float64(n+1)*0.73)
		rect := model3d.NewRect(
			model3d.XYZ(xc - wth/2, yc-depth, 0.01),
			model3d.XYZ(xc + wth/2, yc, stH),
		)
		stairs = append(stairs, rect)
	}

	var cutouts model3d.JoinedSolid
	entryC := model3d.XYZ(0, baseRadius-0.015, 0.092)
	width, height, archH := 0.23, 0.29, 0.10
	doorRect := model3d.NewRect(
		model3d.XYZ(entryC.X-width/2, entryC.Y, entryC.Z),
		model3d.XYZ(entryC.X+width/2, entryC.Y+0.034, entryC.Z+height-archH),
	)
	doorArch := &model3d.Cylinder{
		P1:     entryC.Add(model3d.Y(0.017)).Add(model3d.Z(height-archH/2)),
		P2:     entryC.Add(model3d.Y(0.017+0.008)).Add(model3d.Z(height-archH/2)),
		Radius: width/2 + 0.01,
	}
	doorway := model3d.JoinedSolid{doorRect, doorArch}
	for nb := 0; nb < 3; nb++ {
		x := -0.07 + float64(nb)*0.07
		bar := model3d.NewRect(
			model3d.XYZ(entryC.X+x-0.008, entryC.Y+0.01, 0.13),
			model3d.XYZ(entryC.X+x+0.008, entryC.Y+0.034, 0.31),
		)
		cutouts = append(cutouts, bar)
	}
	cutouts = append(cutouts, doorway)

	for _, spec := range towerSpecs {
		base := model3d.XYZ(spec.dist*math.Cos(spec.theta), spec.dist*math.Sin(spec.theta), baseThick)
		h := base.Add(model3d.Z(spec.height*0.62 + 0.09))
		if spec.octagon {
			cutouts = append(cutouts, model3d.NewRect(
				h.Add(model3d.XYZ(-0.03, -0.009, -0.14)),
				h.Add(model3d.XYZ(0.03, 0.009, 0.14)),
			))
		} else {
			cutouts = append(cutouts,
				&model3d.Sphere{
					Center: h,
					Radius: 0.045 + 0.005*rand.Float64(),
				})
		}
	}
	for m := 0; m < 6; m++ {
		th := float64(m)*math.Pi/3 + 0.15
		d := keepRadius * 0.82
		p := model3d.XYZ(math.Cos(th)*d, math.Sin(th)*d, baseThick + keepHeight*0.46)
		rct := model3d.NewRect(
			p.Add(model3d.XYZ(-0.018, -0.004, -0.083)),
			p.Add(model3d.XYZ(0.018, 0.004, 0.083)),
		)
		cutouts = append(cutouts, rct)
	}

	balcenter := model3d.XYZ(keepRadius*math.Cos(math.Pi/2), keepRadius*math.Sin(math.Pi/2), baseThick + keepHeight*0.38)
	balcony := &model3d.Cylinder{
		P1:     balcenter.Add(model3d.Z(-0.02)),
		P2:     balcenter.Add(model3d.Z(0.031)),
		Radius: keepRadius * 0.30,
	}
	cutouts = append(cutouts, balcony)

	for i := -1; i <= 1; i++ {
		pw := 0.09+0.03*float64(i%2)
		arch := model3d.JoinedSolid{
			model3d.NewRect(
				model3d.XYZ(float64(i)*0.4-pw/2, -baseRadius+0.26, 0.014),
				model3d.XYZ(float64(i)*0.4+pw/2, -baseRadius+0.26+0.03, 0.094),
			),
			&model3d.Cylinder{
				P1: model3d.XYZ(float64(i)*0.4, -baseRadius+0.26+0.017, 0.084),
				P2: model3d.XYZ(float64(i)*0.4, -baseRadius+0.26+0.037, 0.084),
				Radius: pw/2,
			},
		}
		cutouts = append(cutouts, arch)
	}
	for i := 0; i < 9; i++ {
		th := float64(i) * 2 * math.Pi / 9
		if rand.Float64() < 0.60 {
			cutouts = append(cutouts,
				ShellImprint(
					model3d.XYZ((baseRadius-0.09)*math.Cos(th+0.24), (baseRadius-0.09)*math.Sin(th+0.24), 0.06 + 0.04*math.Sin(th*2.2)),
					0.033, 0.011, false),
			)
		}
	}
	for i := 0; i < 7; i++ {
		t := rand.Float64()*2*math.Pi
		wr := baseRadius*0.72
		z := baseThick + 0.17 + 0.027*rand.Float64()
		cutouts = append(cutouts,
			ShellImprint(
				model3d.XYZ(wr*math.Cos(t), wr*math.Sin(t), z), 0.026, 0.009, false))
	}

	var sandDimples model3d.JoinedSolid
	for i := 0; i < 60; i++ {
		r := baseRadius * (0.27 + 0.68*rand.Float64())
		theta := rand.Float64() * 2 * math.Pi
		z := baseThick * (rand.Float64()*0.8 + 0.05)
		center := model3d.XYZ(r*math.Cos(theta), r*math.Sin(theta), z)
		sphere := &model3d.Sphere{
			Center: center,
			Radius: 0.017 + 0.013*rand.Float64(),
		}
		sandDimples = append(sandDimples, sphere)
	}

	var drips model3d.JoinedSolid
	for _, tower := range towerSpecs {
		angle := tower.theta
		r := tower.dist + tower.radius*0.77
		basePt := model3d.XYZ(r*math.Cos(angle), r*math.Sin(angle), tower.height+baseThick+0.14)
		drips = append(drips, DripSpire(basePt, 0.10+rand.Float64()*0.08, tower.radius*0.21, 0.025))
	}

	var flags model3d.JoinedSolid
	flagColorsRGB := []render3d.Color{
		render3d.NewColorRGB(0.87, 0.22, 0.19),
		render3d.NewColorRGB(1.00, 0.96, 0.23),
		render3d.NewColorRGB(0.19, 0.47, 0.94),
	}
	type flagSpec struct {
		base   model3d.Coord3D
		height float64
		width  float64
		colorIdx int
		phase  float64
	}
	flagSpecs := []flagSpec{
		{keepBase.Add(model3d.Z(keepHeight+0.33)), 0.19, 0.18, 0, 0.2},
		{
			model3d.XYZ(
				towerSpecs[0].dist*math.Cos(towerSpecs[0].theta),
				towerSpecs[0].dist*math.Sin(towerSpecs[0].theta),
				baseThick+towerSpecs[0].height+0.16),
			0.13, 0.13, 1, 0.5,
		},
		{
			model3d.XYZ(
				towerSpecs[2].dist*math.Cos(towerSpecs[2].theta),
				towerSpecs[2].dist*math.Sin(towerSpecs[2].theta),
				baseThick+towerSpecs[2].height+0.14),
			0.13, 0.13, 2, -0.3,
		},
	}
	for _, sp := range flagSpecs {
		flags = append(flags, WavyFlag(sp.base, sp.height, sp.width, sp.colorIdx, sp.phase))
	}

	positive := model3d.JoinedSolid{
		baseMinusMoat,
		towers,
		walls,
		battlements,
		bridge,
		stairs,
		drips,
		flags,
	}
	negative := model3d.JoinedSolid{
		cutouts,
		sandDimples,
	}

	finalSolid := model3d.Subtract(positive, negative)

	delta := 0.02
	mesh, points := model3d.DualContourInterior(finalSolid, delta, true, false)
	coordTree := model3d.NewCoordTree(points)

	sandMain := render3d.NewColorRGB(0.93, 0.88, 0.70)
	sandShadow := render3d.NewColorRGB(0.76, 0.65, 0.52)
	sandWarm := render3d.NewColorRGB(1.00, 0.93, 0.83)
	sandBright := render3d.NewColorRGB(0.99, 0.97, 0.87)
	shellCol := render3d.NewColorRGB(0.99, 0.92, 0.93)

	return mesh, func(c model3d.Coord3D) render3d.Color {
		interior := coordTree.NearestNeighbor(c)
		bump := sandBump(interior.X, interior.Y, interior.Z)
		sand := sandMain.Add(render3d.NewColorRGB(0.11, 0.07, 0.028).Scale(bump))

		if interior.Z > keepHeight+baseThick*0.8 {
			sand = sandBright
		} else if interior.Z > baseThick+0.39 && bump > 0.03 {
			sand = sandWarm.Add(render3d.NewColorRGB(0.04, 0.06, 0.03).Scale(bump))
		}
		if interior.Z < 0.025 && math.Hypot(interior.X, interior.Y) > baseRadius-0.13 {
			return sandShadow
		}
		if interior.Y > baseRadius-0.14 && interior.Z < 0.28 {
			return sandBright.Add(render3d.NewColorRGB(0.09, 0.07, 0.01))
		}
		if interior.Z > baseThick+0.4 && bump > 0.05 {
			return sandBright.Add(render3d.NewColorRGB(0.05, 0.05, 0.00))
		}
		for i := 0; i < 9; i++ {
			th := float64(i)*2*math.Pi/9
			pos := model3d.XYZ((baseRadius-0.09)*math.Cos(th+0.24), (baseRadius-0.09)*math.Sin(th+0.24), 0.06+0.04*math.Sin(th*2.2))
			if c.Dist(pos) < 0.035 {
				return shellCol
			}
		}
		for _, sp := range flagSpecs {
			flagCtr := sp.base.Add(model3d.Z(sp.height))
			if c.Dist(flagCtr) < 0.16 {
				return flagColorsRGB[sp.colorIdx]
			}
		}
		return sand
	}
}
