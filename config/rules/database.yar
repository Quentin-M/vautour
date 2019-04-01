rule db_create_user
{
    meta:
        description = "Contains a privileged user creation in a database"
        author = "@KevTheHermit"
        score = 10

    strings:
        $a = "GRANT ALL PRIVILEGES" nocase
        $b = "IDENTIFIED BY" nocase
        $c = "GRANT SELECT" nocase
        $d = "CREATE USER" nocase

    condition:
        2 of them
}