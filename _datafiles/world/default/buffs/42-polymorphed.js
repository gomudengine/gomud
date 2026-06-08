/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    actor.SetAdjective("polymorphed", true);
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    if (triggersLeft == 5) {
        SendUserMessage(actor.UserId(), 'You feel the polymorph weakening...');
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
    SendUserMessage(actor.UserId(), 'The polymorph fades and your body shifts back to normal.');
    SendRoomMessage(actor.GetRoomId(),
        actor.GetCharacterName(true) + ' shimmers and reverts to their true form.',
        actor.UserId());
    actor.SetAdjective("polymorphed", false);
}
