
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You feel warm inside. You feel that you could take on even the harshest winter weather.');
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your inner warmth subsides.');
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' shakes off the haze and the glaze in their eyes fades away.', actor.UserId());
}
