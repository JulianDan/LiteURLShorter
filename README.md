# Lite URL Shorter

## Intro

This is a lightweight URL shorter made by Go.

It doesn't need a database, all data is stored in JSON files.  
Of course, this also means it's performance is not very high.  
On my computer, the insertion and query time for a database containing 20,000 URL records(~700KB) are about 40ms and 20ms.

## Preparation

You should create `data.json` and `user.json` in the work dir.

**data.json**:

```JSON
{}
```

**user.json**:

```JSON
{
    "username": "SHA-256 Of Your Password" 
}
```

You can add many users, but it's useless because we doesn't have a permission management system.

## API

We have 4 APIs for CRUD(create, read, update, delete).

The following table lists the details of these APIs:

| Path          | Method | Data Type  | Data Required |
| ------        | ------ |   ------   | ---- |
| /short_url   | GET     |     No     |  No  |
| /short_url   | POST    |    Form    |  long_url, name, pwd  |
| /short_url   | PATCH   |    Form    |  long_url, name, pwd  |
| /short_url   | DELETE  | URL Params |  name, pwd  |

`name` is the username and `pwd` is the password.
By default, the requests from browser(GET method) will go to first API, if the short URL was found in the database, it will returned a 307 (Temporary redirect)

Obviously, the first API doesn't require an authentication.
And the remaining three APIs require username and password as parameters to auth.
