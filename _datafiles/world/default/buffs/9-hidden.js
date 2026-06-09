
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You feel sneaky.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' disappears into the shadows.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You no longer feel sneaky.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' emerges from the shadows.', actor.UserId());
}
