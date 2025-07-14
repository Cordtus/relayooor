#!/usr/bin/env node

/**
 * Style Audit Script
 * Identifies hardcoded colors, inconsistent spacing, and other style issues
 */

const fs = require('fs')
const path = require('path')
const glob = require('fast-glob')

// Patterns to detect
const patterns = {
  hardcodedColors: /(?:text|bg|border|ring)-(?:gray|red|blue|green|yellow|orange|purple|pink|indigo)-\d{2,3}/g,
  inlineStyles: /style="[^"]+"/g,
  oldButtonClasses: /class="[^"]*(?:px-4 py-2|bg-blue-600|hover:bg-blue-700)[^"]*"/g,
  inconsistentSpacing: /(?:p-[0-9]|m-[0-9]|gap-[0-9]|space-[xy]-[0-9])/g,
  oldCardPattern: /bg-white.*rounded-lg.*shadow/g,
  directColorText: /text-(?:gray-[456789]00|black|white)(?!\w)/g
}

// Files to scan
const filesToScan = [
  'src/**/*.vue',
  'src/**/*.ts',
  'src/**/*.js'
]

async function scanFiles() {
  const files = await glob(filesToScan, { cwd: __dirname + '/..' })
  const issues = []

  for (const file of files) {
    const content = fs.readFileSync(path.join(__dirname, '..', file), 'utf-8')
    const fileIssues = scanContent(content, file)
    
    if (fileIssues.length > 0) {
      issues.push({ file, issues: fileIssues })
    }
  }

  return issues
}

function scanContent(content, filename) {
  const issues = []
  const lines = content.split('\n')

  lines.forEach((line, index) => {
    // Skip import statements and comments
    if (line.trim().startsWith('import') || line.trim().startsWith('//') || line.trim().startsWith('*')) {
      return
    }

    // Check for hardcoded colors
    const colorMatches = line.match(patterns.hardcodedColors)
    if (colorMatches) {
      issues.push({
        line: index + 1,
        type: 'hardcoded-color',
        message: `Hardcoded color classes found: ${colorMatches.join(', ')}`,
        suggestion: 'Use semantic color tokens from the design system'
      })
    }

    // Check for inline styles
    if (patterns.inlineStyles.test(line)) {
      issues.push({
        line: index + 1,
        type: 'inline-style',
        message: 'Inline style attribute found',
        suggestion: 'Use Tailwind utility classes or dynamic class bindings'
      })
    }

    // Check for old button patterns
    if (patterns.oldButtonClasses.test(line) && !filename.includes('Button.vue')) {
      issues.push({
        line: index + 1,
        type: 'old-button',
        message: 'Old button styling pattern found',
        suggestion: 'Use the Button component instead'
      })
    }

    // Check for inconsistent spacing
    const spacingMatches = line.match(patterns.inconsistentSpacing)
    if (spacingMatches) {
      const uniqueSpacing = [...new Set(spacingMatches)]
      if (uniqueSpacing.some(s => !isStandardSpacing(s))) {
        issues.push({
          line: index + 1,
          type: 'inconsistent-spacing',
          message: `Non-standard spacing found: ${uniqueSpacing.join(', ')}`,
          suggestion: 'Use standard spacing: p-card, space-y-section, gap-card'
        })
      }
    }

    // Check for old card patterns
    if (patterns.oldCardPattern.test(line) && !filename.includes('Card.vue')) {
      issues.push({
        line: index + 1,
        type: 'old-card',
        message: 'Old card pattern found',
        suggestion: 'Use class="card-base" instead'
      })
    }
  })

  return issues
}

function isStandardSpacing(spacing) {
  const standardValues = ['2', '3', '4', '5', '6', '8', 'card', 'section']
  const value = spacing.split('-').pop()
  return standardValues.includes(value)
}

function generateReport(allIssues) {
  console.log('# Style Audit Report\n')
  console.log(`Total files scanned: ${allIssues.length}`)
  
  const issueTypes = {}
  let totalIssues = 0

  allIssues.forEach(({ file, issues }) => {
    issues.forEach(issue => {
      totalIssues++
      issueTypes[issue.type] = (issueTypes[issue.type] || 0) + 1
    })
  })

  console.log(`Total issues found: ${totalIssues}\n`)
  
  console.log('## Issue Summary')
  Object.entries(issueTypes).forEach(([type, count]) => {
    console.log(`- ${type}: ${count} occurrences`)
  })

  console.log('\n## Files with Issues\n')
  
  allIssues.forEach(({ file, issues }) => {
    console.log(`### ${file} (${issues.length} issues)`)
    issues.forEach(issue => {
      console.log(`- Line ${issue.line}: ${issue.message}`)
      console.log(`  Suggestion: ${issue.suggestion}`)
    })
    console.log('')
  })

  // Generate fix suggestions
  console.log('## Quick Fix Commands\n')
  console.log('To help fix these issues:')
  console.log('1. Replace hardcoded colors with semantic tokens')
  console.log('2. Convert inline styles to utility classes')
  console.log('3. Replace old button patterns with Button component')
  console.log('4. Standardize spacing to use design tokens')
  console.log('5. Update card patterns to use card-base class')
}

// Run the audit
async function main() {
  console.log('Running style audit...\n')
  const issues = await scanFiles()
  generateReport(issues)
}

main().catch(console.error)