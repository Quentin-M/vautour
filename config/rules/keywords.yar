rule keyword
{
    meta:
        description = "Contains a keyword"
        author = "@KevTheHermit"
        score = 5

    strings:
        $tango_down = "TANGO DOWN" wide ascii nocase
        $antisec = "antisec" wide ascii nocase
        $hacked = "hacked by" wide ascii nocase
        $onion_url = /.*.\.onion/
        $nmap_scan = "Nmap scan report for" wide ascii nocase
        $enabled_sec = "enable secret" wide ascii nocase
        $enable_pass = "enable password" wide ascii nocase
    condition:
        any of them
}

rule dox
{
    meta:
        description = "Contains a dox"
        author = "@KevTheHermit"
        score = 5

    strings:
        $dox = "DOX" wide ascii nocase fullword
        $keyword1 = "name" wide ascii nocase
        $keyword2 = "dob" wide ascii nocase
        $keyword3 = "age" wide ascii nocase
        $keyword4 = "password" wide ascii nocase
        $keyword5 = "email" wide ascii nocase
    condition:
        $dox and 3 of ($keyword*)
}