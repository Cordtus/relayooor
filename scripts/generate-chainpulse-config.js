#!/usr/bin/env node

/**
 * Generate chainpulse configuration from nodes.toml
 * Usage: node generate-chainpulse-config.js [output-file]
 */

const fs = require('fs')
const path = require('path')
const toml = require('@iarna/toml')

// Load nodes configuration
const nodesPath = path.join(__dirname, '..', 'config', 'nodes.toml')
const nodesContent = fs.readFileSync(nodesPath, 'utf-8')
const nodesConfig = toml.parse(nodesContent)

// Generate chainpulse configuration
function generateChainpulseConfig(selectedChains = null) {
  let output = '# Auto-generated chainpulse configuration from nodes.toml\n'
  output += `# Generated on ${new Date().toISOString()}\n\n`
  
  output += '[database]\n'
  output += 'path = "/data/chainpulse.db"\n\n'
  
  output += '[metrics]\n'
  output += 'enabled = true\n'
  output += 'port = 3001\n'
  output += 'stuck_packets = true\n'
  output += 'populate_on_start = true\n\n'
  
  // Filter chains
  const chains = Object.entries(nodesConfig.chains)
    .filter(([chainId, node]) => {
      // If specific chains are selected, only include those
      if (selectedChains && selectedChains.length > 0) {
        return selectedChains.includes(chainId)
      }
      // Otherwise include all healthy and active chains
      return node.healthy && node.active
    })
  
  // Add chains to config
  for (const [chainId, node] of chains) {
    output += `# ${node.name}\n`
    output += `[chains.${chainId}]\n`
    output += `url = "${node.ws}"\n`
    output += `comet_version = "${node.comet_version}"\n\n`
  }
  
  return output
}

// Main execution
const args = process.argv.slice(2)
const outputFile = args[0]

// Example: generate config for specific chains
// node generate-chainpulse-config.js --chains cosmoshub-4,osmosis-1,neutron-1
const chainsArg = args.find(arg => arg.startsWith('--chains='))
const selectedChains = chainsArg ? chainsArg.split('=')[1].split(',') : null

const config = generateChainpulseConfig(selectedChains)

if (outputFile && !outputFile.startsWith('--')) {
  fs.writeFileSync(outputFile, config)
  console.log(`Chainpulse configuration written to ${outputFile}`)
} else {
  console.log(config)
}

// Also show summary
const chains = Object.entries(nodesConfig.chains)
console.error('\nSummary:')
console.error(`Total chains: ${chains.length}`)
console.error(`Healthy chains: ${chains.filter(([_, n]) => n.healthy).length}`)
console.error(`Active chains: ${chains.filter(([_, n]) => n.active).length}`)
console.error(`Healthy & Active: ${chains.filter(([_, n]) => n.healthy && n.active).length}`)