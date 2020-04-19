# go-snoonotes
SnooNotes API Wrapper written in Go!

## Usage

### Authentication
To authenticate a user, you need their reddit username and *SnooNotes User Key*. The User Key can be retrieved from the [SnooNotes website](https://snoonotes.com/#!/userkey).

The package is built with multi-user support in mind, and automatically keeps track of signed in users and their OAuth tokens for you. To auth a new user, simply use:
```golang
if err := snoonotes.Auth(username, snoonotesUserKey); err != nil {
    panic(err)
}
```

Calls to most functions (like `Get`, `Add` etc.) always require the reddit username of a previously authed user to perform the requests **as**.

### Getting & Creating Notes

After authentication, you can make requests to SnooNotes on behalf of the user. To return all notes about the user "RedditUser", simply use `Get`:
```golang
notes, err := snoonotes.Get("politics", username, "RedditUser")
if err != nil {
    panic(err)
}
if notes == nil {
    fmt.Println("No notes found!")
} else {
    for _, note := range *notes {
        fmt.Println(note.Message)
    }
}
```

The first argument (in the example "politics") filters the note by a specific subreddit. You can leave it empty to get notes for all subreddits the user has access to.


To create a new note, initialize a `NewNote` struct and pass it to `Add`:
```golang
note := snoonotes.NewNote{
    NoteTypeID:        1234,
    SubName:           "politics",
    Message:           "User was nice today :)",
    AppliesToUsername: "RedditUser",
    URL:               "https://redd.it/abcdef",
}
if err := snoonotes.Add(username, note); err != nil {
    panic(err)
}
```

### Handling Note Categories

SnooNotes allows you to set different categories for notes (Ban Note, Permaban Note, Misc Note etc.). These categories can be configured on a per-subreddit basis. The simplest way to map a `NoteTypeID` to something useful is `GetNoteTypeMap`:

```golang
types, err := snoonotes.GetNoteTypeMap(username, "politics")
if err != nil {
    panic(err)
}
notes, err := sn.Get("politics", username, "RedditUser")
if err != nil {
    panic(err)
}
for _, note := range *notes {
    noteType, ok := types[note.NoteTypeID];
    if !ok {
        fmt.Println("encountered unknown NoteTypeID - this shouldn't happen!")
        continue
    }
    fmt.Printf("Note Category: %s - Note Text: %s - Note created by: %s\n", noteType.DisplayName, note.Message, note.Submitter)
}
```

To get more details, use the `GetNoteTypes` and `GetConfig` methods. The config fetched from the SnooNotes API is cached for 24 hours, as it rarely changes.