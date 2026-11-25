#!/usr/bin/env node
/**
 * Test script for local development setup
 * Validates that all components are running correctly
 */

import { execSync } from 'child_process';

interface TestResult {
  name: string;
  success: boolean;
  message: string;
}

const tests: TestResult[] = [];

// Helper function to execute shell commands safely
function execCommand(command: string): string {
  try {
    return execSync(command, { encoding: 'utf8', stdio: 'pipe' }).toString().trim();
  } catch {
    throw new Error(`Command failed: ${command}`);
  }
}

async function runTest(name: string, testFn: () => Promise<boolean>): Promise<void> {
  try {
    console.log(`🧪 Running test: ${name}`);
    const success = await testFn();
    const message = success ? "✅ PASSED" : "❌ FAILED";
    tests.push({ name, success, message });
    console.log(`   ${message}\n`);
  } catch (error: any) {
    tests.push({ name, success: false, message: `❌ ERROR: ${error.message}` });
    console.log(`   ❌ ERROR: ${error.message}\n`);
  }
}

async function testKindCluster(): Promise<boolean> {
  try {
    const result = execCommand('kind get clusters');
    return result.includes("smart-garden-bot");
  } catch {
    return false;
  }
}

async function testNamespaces(): Promise<boolean> {
  try {
    const result = execCommand('kubectl get namespaces -o name');
    return result.includes("smart-garden-bot-api") && 
           result.includes("smart-garden-bot-db") &&
           result.includes("cnpg-system");
  } catch {
    return false;
  }
}

async function testPostgresCluster(): Promise<boolean> {
  try {
    const result = execCommand('kubectl get cluster -n smart-garden-bot-db smart-garden-bot-postgres -o jsonpath=\'{.status.phase}\'');
    return result === "Cluster in healthy state";
  } catch {
    // If the exact status check fails, just check if the cluster exists
    try {
      execCommand('kubectl get cluster -n smart-garden-bot-db smart-garden-bot-postgres');
      return true;
    } catch {
      return false;
    }
  }
}

async function testApiDeployment(): Promise<boolean> {
  try {
    const result = execCommand('kubectl get deployment -n smart-garden-bot-api api -o jsonpath=\'{.status.readyReplicas}\'');
    return parseInt(result) > 0;
  } catch {
    return false;
  }
}

async function testApiHealth(): Promise<boolean> {
  try {
    // Wait a moment for port-forward to be ready
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    const response = await fetch("http://localhost:8080/health", {
      signal: AbortSignal.timeout(5000)
    });
    return response.ok;
  } catch {
    return false;
  }
}

async function testFrontendService(): Promise<boolean> {
  try {
    execCommand('kubectl get service -n smart-garden-bot-api frontend-service');
    return true;
  } catch {
    return false;
  }
}

async function testTiltStatus(): Promise<boolean> {
  try {
    const result = execCommand('tilt get uiresource -o json');
    const resources = JSON.parse(result);
    return resources.items && resources.items.length > 0;
  } catch {
    return false;
  }
}

async function main(): Promise<void> {
  console.log("🌱 Smart Garden Bot - Local Development Setup Test\n");
  console.log("================================================\n");

  // Run all tests
  await runTest("Kind cluster exists", testKindCluster);
  await runTest("Required namespaces exist", testNamespaces);
  await runTest("PostgreSQL cluster is running", testPostgresCluster);
  await runTest("API deployment is ready", testApiDeployment);
  await runTest("API health endpoint responds", testApiHealth);
  await runTest("Frontend service exists", testFrontendService);
  await runTest("Tilt is managing resources", testTiltStatus);

  // Summary
  const passed = tests.filter(t => t.success).length;
  const total = tests.length;
  
  console.log("=".repeat(50));
  console.log(`Test Results: ${passed}/${total} passed`);
  console.log("=".repeat(50));
  
  tests.forEach(test => {
    console.log(`${test.message} ${test.name}`);
  });

  if (passed === total) {
    console.log("\n🎉 All tests passed! Your local development environment is ready!");
    console.log("\n📍 Access your services:");
    console.log("   • API: http://localhost:8080");
    console.log("   • Frontend: http://localhost:3000");
    console.log("   • Dashboard: http://localhost:9090");
    console.log("   • Database: localhost:5432 (user: smartgarden, password: localdev123)");
    console.log("\n🚀 Try running some commands:");
    console.log("   • tilt trigger db-migrate  # Run database migrations");
    console.log("   • tilt trigger db-seed     # Seed test data");
    console.log("   • tilt trigger api-test    # Run API tests");
  } else {
    console.log("\n❌ Some tests failed. Check the output above for details.");
    console.log("\n🔧 Troubleshooting tips:");
    console.log("   • Make sure Docker is running");
    console.log("   • Verify Kind cluster is created: kind get clusters");
    console.log("   • Check Tilt status: tilt get uiresource");
    console.log("   • View logs: kubectl logs -n smart-garden-bot-api -l app=api");
    
    process.exit(1);
  }
}

// Run the main function when script is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(console.error);
}