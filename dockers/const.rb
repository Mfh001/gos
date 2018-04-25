AppNames = {"AuthApp"    => "gos-auth-app",
            "ConnectApp" => "gos-connect-app",
            "GameApp"    => "gos-game-app",
            "WorldApp"   => "gos-world-app"}

AppMappings = {"gos-auth-app" => "-p 3000:3000",
               "gos-connect-app" => "-p 4000:4000",
               "gos-game-app" => "-p 50053:50053",
               "gos-world-app" => "-p 50051:50051 -p 50052:50052"}
