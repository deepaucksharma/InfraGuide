# New Relic Ingest & Instrumentation Landscape - Content Catalog Generator
# This script generates a structured catalog of all content in the book

function Get-ContentSummary {
    param (
        [string]$FilePath
    )
    
    if (Test-Path $FilePath) {
        $content = Get-Content $FilePath -Raw
        
        # Try to extract the first heading
        $headingMatch = [regex]::Match($content, "^# (.+)$", "Multiline")
        $heading = if ($headingMatch.Success) { $headingMatch.Groups[1].Value } else { "(No heading found)" }
        
        # Count diagrams, tables, and code blocks
        $mermaidCount = ([regex]::Matches($content, "```mermaid")).Count
        $tableCount = ([regex]::Matches($content, "\|\s*-+\s*\|")).Count
        $codeBlockCount = ([regex]::Matches($content, "```(yaml|json|sql|nrql)")).Count
        
        # Estimate reading time (average reading speed: 200 words per minute)
        $wordCount = ($content -split '\s+').Count
        $readingTimeMinutes = [math]::Ceiling($wordCount / 200)
        
        return @{
            Title = $heading
            MermaidDiagrams = $mermaidCount
            Tables = $tableCount
            CodeBlocks = $codeBlockCount
            WordCount = $wordCount
            ReadingTimeMinutes = $readingTimeMinutes
        }
    }
    else {
        return @{
            Title = "(File not found)"
            MermaidDiagrams = 0
            Tables = 0
            CodeBlocks = 0
            WordCount = 0
            ReadingTimeMinutes = 0
        }
    }
}

function Format-Time {
    param (
        [int]$Minutes
    )
    
    if ($Minutes -lt 60) {
        return "$Minutes min"
    }
    else {
        $hours = [math]::Floor($Minutes / 60)
        $mins = $Minutes % 60
        return "$hours h $mins min"
    }
}

# Get the root directory of the project
$rootDir = $PSScriptRoot

# Set the output file for the catalog
$outputFile = Join-Path $rootDir "content_catalog.md"

# Initialize the output
$output = @"
# New Relic Ingest & Instrumentation Landscape
## Complete Content Catalog
Generated: $(Get-Date -Format "yyyy-MM-dd")

"@

# Get all section directories, ordered by prefix
$sectionDirs = Get-ChildItem -Path $rootDir -Directory | 
                Where-Object { $_.Name -match "^\d{2}_" } | 
                Sort-Object Name

$totalMermaid = 0
$totalTables = 0
$totalCode = 0
$totalWords = 0
$totalReadingTime = 0

foreach ($sectionDir in $sectionDirs) {
    $sectionName = $sectionDir.Name -replace "^\d{2}_", ""
    $sectionName = $sectionName -replace "_", " "
    
    $output += "## $sectionName`r`n`r`n"
    $output += "| Chapter | Type | Diagrams | Tables | Code Blocks | Word Count | Reading Time |`r`n"
    $output += "|---------|------|----------|--------|-------------|------------|--------------|`r`n"
    
    # Get all markdown files in the section, ordered by prefix
    $chapterFiles = Get-ChildItem -Path $sectionDir.FullName -Filter "*.md" | 
                     Sort-Object Name
    
    $sectionMermaid = 0
    $sectionTables = 0
    $sectionCode = 0
    $sectionWords = 0
    $sectionReadingTime = 0
    
    foreach ($chapterFile in $chapterFiles) {
        $fileName = $chapterFile.Name
        $chapterName = $fileName -replace "^\d{2}_", "" -replace "\.md$", "" -replace "_", " "
        
        $summary = Get-ContentSummary -FilePath $chapterFile.FullName
        
        $output += "| [$chapterName](./$($sectionDir.Name)/$fileName) | $($summary.Title) | $($summary.MermaidDiagrams) | $($summary.Tables) | $($summary.CodeBlocks) | $($summary.WordCount) | $(Format-Time -Minutes $summary.ReadingTimeMinutes) |`r`n"
        
        $sectionMermaid += $summary.MermaidDiagrams
        $sectionTables += $summary.Tables
        $sectionCode += $summary.CodeBlocks
        $sectionWords += $summary.WordCount
        $sectionReadingTime += $summary.ReadingTimeMinutes
    }
    
    $output += "| **Section Total** | | **$sectionMermaid** | **$sectionTables** | **$sectionCode** | **$sectionWords** | **$(Format-Time -Minutes $sectionReadingTime)** |`r`n`r`n"
    
    $totalMermaid += $sectionMermaid
    $totalTables += $sectionTables
    $totalCode += $sectionCode
    $totalWords += $sectionWords
    $totalReadingTime += $sectionReadingTime
}

$output += @"
## Summary Statistics

- **Total Chapters**: $((Get-ChildItem -Path $rootDir -Filter "*.md" -Recurse | Measure-Object).Count)
- **Total Mermaid Diagrams**: $totalMermaid
- **Total Tables**: $totalTables
- **Total Code Blocks**: $totalCode
- **Total Word Count**: $totalWords
- **Estimated Reading Time**: $(Format-Time -Minutes $totalReadingTime)

"@

# Write the output to the catalog file
$output | Set-Content -Path $outputFile -Encoding UTF8

Write-Host "Content catalog has been generated at: $outputFile"
