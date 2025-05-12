const teacherMobId = 57;

function onDie(mob, room, eventDetails) {

    room.SendText( mob.GetCharacterName(true) + " crumbles to dust." );

    teacherMob = getTeacher(room);

    teacherMob.Command('say You did it! As you can see you gain <ansi fg="experience">experience points</ansi> for combat victories.');
    teacherMob.Command('say Now head <ansi fg="exit">west</ansi> to complete your training.', 2.0);
}


function getTeacher(room) {

    var mobActor = null;

    mobIds = room.GetMobs();
        
    for (var i in mobIds ) {
        mobActor = GetMob(mobIds[i]);
        if ( mobActor.MobTypeId() == teacherMobId ) {
            return mobActor;
        }
    }

    mobActor = room.SpawnMob(teacherMobId);
    mobActor.SetCharacterName(teacherName);

    return mobActor;
}
