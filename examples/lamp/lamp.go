package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

// Rodrigues' rotation formula.
func RotateVec(vec, axis model3d.Coord3D, angle float64) model3d.Coord3D {
	// axis must be normalized!
	c := math.Cos(angle)
	s := math.Sin(angle)
	d := 1 - c
	ax := axis.Normalize()
	v := vec
	return model3d.Coord3D{
		X: (c+ax.X*ax.X*d)*v.X + (ax.X*ax.Y*d-ax.Z*s)*v.Y + (ax.X*ax.Z*d+ax.Y*s)*v.Z,
		Y: (ax.Y*ax.X*d+ax.Z*s)*v.X + (c+ax.Y*ax.Y*d)*v.Y + (ax.Y*ax.Z*d-ax.X*s)*v.Z,
		Z: (ax.Z*ax.X*d-ax.Y*s)*v.X + (ax.Z*ax.Y*d+ax.X*s)*v.Y + (c+ax.Z*ax.Z*d)*v.Z,
	}
}

// Helper for a disk (cylinder-like but short, with rounded corners)
func RoundedDisk(center model3d.Coord3D, radius, thickness, chamfer float64) model3d.Solid {
	flat := &model3d.Cylinder{
		P1:     center.Add(model3d.Z(-thickness / 2)),
		P2:     center.Add(model3d.Z(thickness / 2)),
		Radius: radius - chamfer,
	}
	if chamfer == 0 {
		return flat
	}
	chamferTop := &model3d.Sphere{
		Center: center.Add(model3d.Z(thickness / 2)),
		Radius: chamfer,
	}
	chamferBot := &model3d.Sphere{
		Center: center.Add(model3d.Z(-thickness / 2)),
		Radius: chamfer,
	}
	return model3d.JoinedSolid{flat, chamferTop, chamferBot}
}

// Helper for a capped truncated cone (lamp shade shell)
func TruncatedCone(baseCenter, tipCenter model3d.Coord3D, baseRadius, tipRadius float64) model3d.Solid {
	dir := tipCenter.Sub(baseCenter).Normalize()
	height := tipCenter.Sub(baseCenter).Norm()
	coneSolid := model3d.CheckedFuncSolid(
		baseCenter.Sub(model3d.Ones(baseRadius+0.05)),
		tipCenter.Add(model3d.Ones(baseRadius+0.05)),
		func(c model3d.Coord3D) bool {
			vec := c.Sub(baseCenter)
			along := vec.Dot(dir)
			if along < 0 || along > height {
				return false
			}
			baseLerp := along / height
			rad := (1-baseLerp)*baseRadius + baseLerp*tipRadius
			dist := vec.Sub(dir.Scale(along)).Norm()
			return dist < rad
		},
	)
	// Cap ring at big end
	ringTh := (baseRadius - tipRadius) * 0.22
	capRing := &model3d.Cylinder{
		P1:     baseCenter.Add(dir.Scale(0.04)),
		P2:     baseCenter.Add(dir.Scale(0.04 + ringTh)),
		Radius: baseRadius * 1.04,
	}
	return model3d.JoinedSolid{coneSolid, capRing}
}

func CreateModel() (*model3d.Mesh, toolbox3d.CoordColorFunc) {
	// PARAMETERS
	baseRadius := 0.66
	baseThickness := 0.19
	baseChamfer := 0.042
	baseRimRad := baseRadius * 1.07

	switchRadius := 0.065
	switchHeight := 0.077

	// Arms
	lowerArmLen := 0.95
	upperArmLen := 0.7
	lowerArmR := 0.086
	upperArmR := 0.062

	// Angles:
	lowerArmAngle := math.Pi / 14  // ~13° from vertical towards -X
	upperArmAngle := math.Pi / 2.6 // ~69° from lower arm (up-forward+X)
	upperArmAzimuth := 0.22        // about 12.5° toward +Y

	// Joint & Joint details
	jointR := 0.143
	jointCollarR := jointR * 1.19
	jointCollarTh := 0.052
	boltRadius := 0.04
	boltHeight := 0.055

	// Head/lampshade:
	shadeLen := 0.35
	shadeBaseRad := 0.21
	shadeTipRad := 0.13
	shadeTipLen := 0.07 // "lip" at front

	// Bulb & socket
	bulbRad := 0.07
	socketRad := bulbRad * 1.15
	socketTh := 0.065

	// Glass/diffuser
	diffuserTh := 0.022

	// Cable
	cableRad := 0.019

	// BASE
	baseCenter := model3d.Z(baseThickness / 2)
	baseDisk := RoundedDisk(baseCenter, baseRadius, baseThickness, baseChamfer)
	baseRim := &model3d.Torus{
		Center:      model3d.Z(baseChamfer * 0.44),
		Axis:        model3d.Z(1),
		OuterRadius: baseRimRad,
		InnerRadius: 0.034,
	}
	baseTopDisk := &model3d.Cylinder{
		P1:     model3d.Z(baseThickness - 0.021),
		P2:     model3d.Z(baseThickness - 0.005),
		Radius: baseRadius * 0.83,
	}
	switchPos := model3d.XY(baseRadius*0.66, -baseRadius*0.13).Add(model3d.Z(baseThickness))
	switchBtn := &model3d.Cylinder{
		P1:     switchPos,
		P2:     switchPos.Add(model3d.Z(switchHeight)),
		Radius: switchRadius,
	}

	// LOWER ARM
	jointBase := model3d.XYZ(0, 0, baseThickness)
	laDir := model3d.XYZ(-math.Sin(lowerArmAngle), 0, math.Cos(lowerArmAngle))
	lowerArmEnd := jointBase.Add(laDir.Scale(lowerArmLen))
	lowerArm := &model3d.Cylinder{
		P1:     jointBase,
		P2:     lowerArmEnd,
		Radius: lowerArmR,
	}
	// Joint 1 (sphere and collar)
	joint1Center := lowerArmEnd
	joint1 := &model3d.Sphere{Center: joint1Center, Radius: jointR}
	joint1Collar := &model3d.Cylinder{
		P1:     joint1Center.Add(laDir.Scale(-jointCollarTh / 2)),
		P2:     joint1Center.Add(laDir.Scale(jointCollarTh / 2)),
		Radius: jointCollarR,
	}
	// Bolt head on joint (axis orthogonal to arm direction)
	boltDir := laDir.Cross(model3d.Y(1)).Normalize()
	joint1Bolt := &model3d.Cylinder{
		P1:     joint1Center.Add(boltDir.Scale(-boltHeight / 2)),
		P2:     joint1Center.Add(boltDir.Scale(boltHeight / 2)),
		Radius: boltRadius,
	}
	joint1BoltHead := &model3d.Sphere{
		Center: joint1Center.Add(boltDir.Scale(boltHeight / 2)),
		Radius: boltRadius * 1.11,
	}

	// UPPER ARM
	uaBase := joint1Center
	axis1 := laDir
	// Rotate axis by upperArmAngle around Y, then by azimuth around axis
	uaTempDir := RotateVec(axis1, model3d.Y(1), upperArmAngle)
	armAxis := RotateVec(uaTempDir, axis1, upperArmAzimuth).Normalize()
	upperArmEnd := uaBase.Add(armAxis.Scale(upperArmLen))
	upperArm := &model3d.Cylinder{
		P1:     uaBase,
		P2:     upperArmEnd,
		Radius: upperArmR,
	}
	// Joint 2
	joint2Center := upperArmEnd
	joint2 := &model3d.Sphere{Center: joint2Center, Radius: jointR * 0.95}
	joint2Collar := &model3d.Cylinder{
		P1:     joint2Center.Add(armAxis.Scale(-jointCollarTh / 2)),
		P2:     joint2Center.Add(armAxis.Scale(jointCollarTh / 2)),
		Radius: jointCollarR * 0.91,
	}
	joint2BoltDir := armAxis.Cross(model3d.Y(1)).Normalize()
	joint2Bolt := &model3d.Cylinder{
		P1:     joint2Center.Add(joint2BoltDir.Scale(-boltHeight / 2)),
		P2:     joint2Center.Add(joint2BoltDir.Scale(boltHeight / 2)),
		Radius: boltRadius,
	}
	joint2BoltHead := &model3d.Sphere{
		Center: joint2Center.Add(joint2BoltDir.Scale(boltHeight / 2)),
		Radius: boltRadius * 1.09,
	}

	// LAMP HEAD
	// Lamp shade orientation: 33 degrees down from arm axis
	shadeDir := RotateVec(armAxis, armAxis.Cross(model3d.Y(1)), -math.Pi/9).Normalize()
	shadeBase := joint2Center.Add(shadeDir.Scale(jointR + 0.032))
	shadeTip := shadeBase.Add(shadeDir.Scale(shadeLen))
	shadeShell := TruncatedCone(shadeBase, shadeTip, shadeBaseRad, shadeTipRad)
	// "Lip" at front
	shadeLip := &model3d.Cylinder{
		P1:     shadeTip,
		P2:     shadeTip.Add(shadeDir.Scale(shadeTipLen)),
		Radius: shadeTipRad * 1.06,
	}

	// Bulb & socket (bulb inside the lampshade)
	bulbCenter := shadeBase.Add(shadeDir.Scale(shadeLen * 0.69))
	bulb := &model3d.Sphere{Center: bulbCenter, Radius: bulbRad}
	bulbSocket := &model3d.Cylinder{
		P1:     bulbCenter.Sub(shadeDir.Scale(socketTh / 2)),
		P2:     bulbCenter.Add(shadeDir.Scale(socketTh / 2)),
		Radius: socketRad,
	}
	// Diffuser/cover (disk near tip)
	diffuserCenter := shadeTip.Add(shadeDir.Scale(shadeTipLen * 0.57))
	diffuser := &model3d.Cylinder{
		P1:     diffuserCenter.Add(shadeDir.Scale(-diffuserTh / 2)),
		P2:     diffuserCenter.Add(shadeDir.Scale(diffuserTh / 2)),
		Radius: shadeTipRad * 0.98,
	}

	// CABLE (run a slightly curved cylinder from shade base,
	// down arm, then angles near base)
	cableA := shadeBase.Add(shadeDir.Scale(-cableRad * 1.1)).Add(model3d.Y(cableRad))
	cableB := joint2Center.Add(model3d.Y(upperArmR * -1.02))
	cableSeg1 := &model3d.Cylinder{
		P1:     cableA,
		P2:     cableB,
		Radius: cableRad,
	}
	cableC := joint1Center.Add(model3d.Y(-lowerArmR * 1.06))
	cableSeg2 := &model3d.Cylinder{
		P1:     cableB,
		P2:     cableC,
		Radius: cableRad,
	}
	baseCableEntry := model3d.XY(baseRadius*(-0.53), 0).Add(model3d.Z(baseThickness - 0.08))
	cableSeg3 := &model3d.Cylinder{
		P1:     cableC,
		P2:     baseCableEntry,
		Radius: cableRad,
	}

	// Compose all
	solid := model3d.JoinedSolid{
		baseDisk, baseRim, baseTopDisk, switchBtn,
		lowerArm, joint1, joint1Collar, joint1Bolt, joint1BoltHead,
		upperArm, joint2, joint2Collar, joint2Bolt, joint2BoltHead,
		shadeShell, shadeLip, bulb, bulbSocket, diffuser,
		cableSeg1, cableSeg2, cableSeg3,
	}

	delta := 0.022
	mesh, points := model3d.MarchingCubesInterior(solid, delta, 8)

	// ---- COLORING -----
	cf := toolbox3d.JoinedSolidCoordColorFunc(
		points,
		// BASE
		baseDisk, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.72, 0.78, 0.87)
		},
		baseRim, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.58, 0.62, 0.69)
		},
		baseTopDisk, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.56, 0.59, 0.66)
		},
		switchBtn, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.98, 0.39, 0.13)
		},
		// LOWER ARM + JOINT
		lowerArm, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.81, 0.83, 0.88)
		},
		joint1, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.18, 0.18, 0.22)
		},
		joint1Collar, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.38, 0.38, 0.43)
		},
		joint1Bolt, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.41, 0.42, 0.37)
		},
		joint1BoltHead, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.2, 0.2, 0.22)
		},
		// UPPER ARM + JOINT
		upperArm, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.82, 0.84, 0.92)
		},
		joint2, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.18, 0.18, 0.24)
		},
		joint2Collar, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.33, 0.34, 0.37)
		},
		joint2Bolt, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.42, 0.41, 0.33)
		},
		joint2BoltHead, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.19, 0.21, 0.21)
		},
		// LAMP HEAD - use blue shade
		shadeShell, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.44, 0.55, 0.88)
		},
		shadeLip, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.30, 0.37, 0.66)
		},
		bulb, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(1.0, 0.96, 0.68)
		},
		bulbSocket, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.47, 0.47, 0.53)
		},
		diffuser, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.98, 0.97, 0.95)
		},
		// CABLE
		cableSeg1, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.07, 0.05, 0.07)
		},
		cableSeg2, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.07, 0.05, 0.07)
		},
		cableSeg3, func(c model3d.Coord3D) render3d.Color {
			return render3d.NewColorRGB(0.07, 0.05, 0.07)
		},
	)

	return mesh, cf
}
