
    请求-验证
path=SessionAuthParams
params={AccountId: string, Token: string}
    
    回复-验证
path=SessionAuthResponse
params={Success: bool}
    
    请求-心跳
path=HeartbeatParams
params={Alive: bool}
    
    请求-穿装备
path=EquipLoadParams
params={PlayerID: string, EquipId: string, HeroId: string}
    
    回复-穿装备
path=EquipLoadResponse
params={PlayerID: string, EquipId: string, Level: uint32}
    
    请求-卸装备
path=EquipUnLoadParams
params={PlayerID: string, EquipId: string, HeroId: string}
    
    回复-卸装备
path=EquipUnLoadResponse
params={PlayerID: string, EquipId: string, Level: uint32}
    
    请求-加入房间
path=RoomJoinParams
params={RoomId: string, PlayerId: string}
    
    回复-加入房间
path=RoomJoinResponse
params={Success: bool}
    
    通知-加入房间
path=RoomJoinNotice
params={RoomId: string, NewPlayerId: string}
    