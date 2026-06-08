
/**
 * Called when a user enters the room.
 * @param {ActorObject} user - The user entering the room.
 * @param {RoomObject} room - The room being entered.
 * @returns {boolean} Return false to suppress the automatic look.
 */
function onEnter(user, room) {
    
    user.SendText('  <ansi fg="red">To get started, type <ansi fg="command">look</ansi> or <ansi fg="command">start</ansi>.</ansi>');
    user.SendText('');

    return true;
}
