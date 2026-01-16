
// Invoked when the buff is first applied to the player.
function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'You notice that the world looks more vibrant and alive.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' seems to be distracted.', actor.UserId());
}

// Invoked every time the buff is triggered (see roundinterval)
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'The colors and sounds of the world captivate you.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' stares into the distance.', actor.UserId());
}

// Invoked when the buff has run its course.
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(),     'The world starts to look a little more normal.');
    SendRoomMessage(actor.GetRoomId(),  actor.GetCharacterName(true)+' seems sober again.', actor.UserId());
}
