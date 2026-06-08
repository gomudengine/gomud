
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'A warm glow surrounds you.');
    SendRoomMessage(actor.GetRoomId(),  'A warm glow surrounds '+actor.GetCharacterName(true)+ '.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'Your glowing fades away.' );
    SendRoomMessage(actor.GetRoomId(),  'The glow surrounding '+actor.GetCharacterName(true)+ ' fades away.', actor.UserId());
}
