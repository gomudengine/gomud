const teacherMobId = 57;

/**
 * Called when the mob dies.
 * @param {ActorObject} mob - The mob that died.
 * @param {RoomObject} room - The room where the mob died.
 * @param {DieEventDetails} eventDetails - Details about the death event.
 * @returns {boolean} Return true if the event was handled.
 */
function onDie(mob, room, eventDetails) {

    room.SendText( mob.GetCharacterName(true) + " crumbles to dust." );

    var teacherMob = room.GetMob(teacherMobId, true);
    if ( teacherMob != null ) {
        teacherMob.Command('say You did it! As you can see you gain <ansi fg="experience">experience points</ansi> for combat victories.');
        teacherMob.Command('say Now head <ansi fg="exit">west</ansi> to complete your training.', 2.0);
    }
}
