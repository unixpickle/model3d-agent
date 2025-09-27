# model3d-agent

An AI agent for creating 3D models with my [model3d](https://github.com/unixpickle/model3d) framework.

The agent is inspired by [LL3M](https://arxiv.org/pdf/2508.08228) and similar works which control Blender with code, except in this case it uses my particular 3D modeling framework.

In practice, the quality is pretty terrible, but it's neat to see the code that the model produces. Sometimes it manages to spew out hundreds of lines of code that compiles.

# Examples

| Caption                             | Rendering                                        | Code                                |
|-------------------------------------|--------------------------------------------------|-------------------------------------|
| a cactus                            | ![Cactus rendering][examples/cactus/cactus.gif]  | [Code][examples/cactus/cactus.go]   |
| a fancy sand castle                 | ![Sandcastle rendering][examples/sandcastle/sandcastle.gif] | [Code][examples/sandcastle/sandcastle.go] |
| a red five piece drum set with cymbals | ![Drums rendering][examples/drums/drums.gif]    | [Code][examples/drums/drums.go]     |
| an adjustable table lamp            | ![Lamp rendering][examples/lamp/lamp.gif]        | [Code][examples/lamp/lamp.go]       |

