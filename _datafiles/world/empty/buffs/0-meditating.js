// 
// buff zero (0) is a special buff that when naturally expires, 
// will remove the player from the game without zombie status.
//

// Invoked when the buff is first applied to the player.
/**
 * Called when the buff is first applied to the actor.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),    'You sit down and begin your meditation.' );
    SendUserMessage(actor.UserId(),    'Your meditation must complete without interruption to quit gracefully.');
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' sits down a begins to meditate.', actor.UserId());
}

// Invoked every time the buff is triggered (see roundinterval)
/**
 * Called each round while the buff is active.
 * @param {ActorObject} actor - The actor the buff is applied to.
 * @param {number} triggersLeft - How many trigger rounds remain.
 * @returns {void}
 */
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You continue your meditation. <ansi bg="blue"> *' + triggersLeft + ' rounds left* </ansi>' );
    SendRoomMessage(actor.GetRoomId(),   actor.GetCharacterName(true)+' continues meditating.', actor.UserId() );
}
