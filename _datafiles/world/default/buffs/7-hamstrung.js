
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'ve been hamstrung!');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' has been hamstrung!', actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You\'re hamstrung!');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is hamstrung!', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'Your leg heals enough that you can fight again.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is no longer hamstrung.', actor.UserId());
}
