package main

import (
    "github.com/oakmound/oak/v4"
    "github.com/oakmound/oak/v4/scene"
    "concurrentec2/scenes"
)

func maina() {
    oak.AddScene("main", scene.Scene{
        Start: escenas.MainScene,
    })

    oak.Init("main", func(c oak.Config) (oak.Config, error) {
        c.BatchLoad = true
        c.Assets.ImagePath = "assets/images"
        return c, nil
    })
}
