rule email_filter
{
    meta:
        description = "Contains several e-mail addresses"
        author = "@kovacsbalu"
        score = 10

    strings:
	    $email_add = /\b[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)*\.[a-zA-Z-]+[\w-]\b/
    condition:
        #email_add > 50
}