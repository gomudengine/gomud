/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your body shifts and transforms!');
    SendRoomMessage(actor.GetRoomId(),
        actor.GetCharacterName(true) + ' transforms before your eyes!',
        actor.UserId());
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    if (triggersLeft == 5) {
        SendUserMessage(actor.UserId(), 'You feel your form beginning to waver...');
    }
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    actor.RevertFormChange();
    SendUserMessage(actor.UserId(), 'Your body shifts back to its original form.');
    SendRoomMessage(actor.GetRoomId(),
        actor.GetCharacterName(true) + ' reverts to their true form!',
        actor.UserId());
}
