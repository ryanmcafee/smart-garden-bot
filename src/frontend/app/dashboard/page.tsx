'use client';

import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';

// Simple placeholder components
function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg border p-6 ${className}`}>
      {children}
    </div>
  );
}

function CardHeader({ children }: { children: React.ReactNode }) {
  return <div className="mb-4">{children}</div>;
}

function CardTitle({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return <h3 className={`text-lg font-semibold ${className}`}>{children}</h3>;
}

function CardDescription({ children }: { children: React.ReactNode }) {
  return <p className="text-sm text-gray-600 dark:text-gray-400">{children}</p>;
}

function CardContent({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return <div className={className}>{children}</div>;
}

export default function DashboardPage() {
  // Mock data query
  const {
    data: mockData,
    isLoading,
  } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      return {
        gardens: 2,
        temperature: 24,
        soilMoisture: 75,
        nextWatering: '2h'
      };
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Smart Garden Dashboard</h1>
          <p className="text-gray-600 dark:text-gray-400">
            Monitor and control your garden automation system
          </p>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Active Gardens</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{mockData?.gardens || 0}</div>
            <p className="text-xs text-gray-600 dark:text-gray-400">
              Total managed gardens
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Current Temperature</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{mockData?.temperature || 0}°C</div>
            <p className="text-xs text-gray-600 dark:text-gray-400">
              Outside temperature
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Soil Moisture</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{mockData?.soilMoisture || 0}%</div>
            <p className="text-xs text-gray-600 dark:text-gray-400">
              Average across zones
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Next Watering</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{mockData?.nextWatering || 'N/A'}</div>
            <p className="text-xs text-gray-600 dark:text-gray-400">
              Vegetable zone
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <Card>
        <CardHeader>
          <CardTitle>Garden Status</CardTitle>
          <CardDescription>
            Your garden automation system is running smoothly
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
              <div>
                <h4 className="font-medium text-green-800 dark:text-green-200">System Online</h4>
                <p className="text-sm text-green-600 dark:text-green-300">All sensors and controls are functioning normally</p>
              </div>
              <div className="w-3 h-3 bg-green-500 rounded-full"></div>
            </div>
            
            <div className="flex items-center justify-between p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <div>
                <h4 className="font-medium text-blue-800 dark:text-blue-200">Weather Integration Active</h4>
                <p className="text-sm text-blue-600 dark:text-blue-300">Monitoring weather conditions for smart watering decisions</p>
              </div>
              <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}