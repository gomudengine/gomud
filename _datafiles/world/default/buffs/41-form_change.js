function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your body shifts and transforms!');
    SendRoomMessage(actor.GetRoomId(),
        actor.GetCharacterName(true) + ' transforms before your eyes!',
        actor.UserId());
}

function onTrigger(actor, triggersLeft) {
    if (triggersLeft == 5) {
        SendUserMessage(actor.UserId(), 'You feel your form beginning to waver...');
    }
}

function onEnd(actor, triggersLeft) {
    actor.RevertFormChange();
    SendUserMessage(actor.UserId(), 'Your body shifts back to its original form.');
    SendRoomMessage(actor.GetRoomId(),
        actor.GetCharacterName(true) + ' reverts to their true form!',
        actor.UserId());
}
