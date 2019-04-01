rule b64_exe
{
    meta:
        description = "Base64'd .exe"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_exe = /\bTV(oA|pB|pQ|qA|qQ|ro)/
        // Double b64 = VFZxUU
    condition:
        $b64_exe at 0
}

rule b64_elf
{
    meta:
        description = "Base64'd ELF"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_elf = "f0VM"
    condition:
        $b64_elf at 0
}

rule b64_zip
{
    meta:
        description = "Base64'd ZIP"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_zip = "UEs"
    condition:
        $b64_zip at 0
}

rule b64_rar
{
    meta:
        description = "Base64'd RAR"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_rar = "UmFy"
    condition:
        $b64_rar at 0
}

rule b64_gzip
{
    meta:
        description = "Base64'd gzip"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_gzip = "H4sI"
    condition:
        $b64_gzip at 0
}

rule b64_url
{
    meta:
        description = "Base64'd URL"
        author = "@KevTheHermit"
        score = 3

    strings:
        $a1 = "aHR0cDov" // http/s
        $a2 = "SFRUUDov" // HTTP/S
        $a3 = "d3d3Lg" // www.
        $a4 = "V1dXLg" // WWW.
        $not1 = "GlobalSign Root CA" nocase
    condition:
        any of ($a*) and not any of ($not*)
}

rule b64_doc
{
    meta:
        description = "Base64'd .doc"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_doc = "0M8R4" // d0cf11
    condition:
        $b64_doc at 0
}

rule b64_rtf
{
    meta:
        description = "Base64'd .rtf"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_rtf = "e1xydGY" // {\rtf
    condition:
        $b64_rtf at 0
}

rule b64_docx
{
    meta:
        description = "Base64'd .docx"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_zip = "UEs"
        $docx1 = "d29yZC9fcmVsc" // word/_rel
        $docx2 = "Zm9udFRhYmxl" // fontTable
        $docx3 = "ZG9jUHJvcHM" // docProps
        $docx4 = "Q29udGVudF9UeXBlcw" // Content_Types
        $docx5 = "c2V0dGluZ3M" //settings
    condition:
        $b64_zip at 0 and 3 of ($docx*)
}

rule b64_xml_doc
{
    meta:
        description = " Base64'd XML doc"
        author = "@KevTheHermit"
        score = 3

    strings:
        $b64_xml = "PD94bWwg"
        $docx1 = "b3BlbmRvY3VtZW50" // opendocument
        $docx2 = "InBhcmFncmFwaCI" // "paragraph"
        $docx3 = "b2ZmaWNlL3dvcmQv" // office/word/
        $docx4 = "RG9jdW1lbnRQcm9wZXJ0aWVz" // DocumentProperties
    condition:
        $b64_xml at 0 and 3 of ($docx*)
}
