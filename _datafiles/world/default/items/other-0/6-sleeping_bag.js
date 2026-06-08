
/**
 * Called when a user issues a use command on the item.
 * @param {ActorObject} user - The user issuing the command.
 * @param {ItemObject} item - The item.
 * @param {RoomObject} room - The room where the command was issued.
 * @returns {boolean} Return true if the command was handled.
 */
function onCommand_use(user, item, room) {
    
    SendUserMessage(user.UserId(), "You unroll the <ansi fg=\"itemname\">"+item.Name()+"</ansi> and hop in.");
    SendRoomMessage(room.RoomId(), user.GetCharacterName(true)+" unrolls their <ansi fg=\"itemname\">"+item.Name()+"</ansi> and crawls inside.", user.UserId());

    user.CancelBuffWithFlag("hidden"); // cancel any hidden buff (most item use should do this if it's noticeable)

    user.GiveBuff(15, "sleep"); // Give the sleeping buff
    
    item.AddUsesLeft(-1); // Decrement the uses left by 1
    item.MarkLastUsed(); // Update the last used round number to current

    return true;
}
