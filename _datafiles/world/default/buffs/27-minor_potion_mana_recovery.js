
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The potion warms you as you drink it down.');
}

/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    manaAmt = actor.AddMana(UtilDiceRoll(1, 5));

    SendUserMessage(actor.UserId(),     'You recover <ansi fg="mana-100">'+String(manaAmt)+' mana</ansi>!');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' is recovery mana from the effects of a potion.', actor.UserId());
}

/**
 * Called when the buff expires or is removed.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The mana potions effect runs out.');
}

