import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import { useGardenStore } from '@/store/garden-store';
import { useNotificationStore } from '@/store/garden-store';
import type { SensorReading } from '@/lib/api-client';

interface UseRealTimeDataOptions {
  enabled?: boolean;
  gardenId?: string;
}

export function useRealTimeData({ enabled = true, gardenId }: UseRealTimeDataOptions = {}) {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;
  const baseReconnectDelay = 1000; // 1 second

  const queryClient = useQueryClient();
  const { updateSensorReading, setWateringStatus } = useGardenStore();
  const { addNotification } = useNotificationStore();

  const connect = () => {
    if (!enabled || !gardenId) return;

    try {
      const eventSource = apiClient.createEventSource(`/stream/garden/${gardenId}`);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        console.log('Real-time connection opened');
        setIsConnected(true);
        setConnectionError(null);
        reconnectAttempts.current = 0;
      };

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          handleRealtimeMessage(data);
        } catch (error) {
          console.error('Failed to parse real-time message:', error);
        }
      };

      eventSource.onerror = (error) => {
        console.error('Real-time connection error:', error);
        setIsConnected(false);
        setConnectionError('Connection lost');
        
        // Attempt reconnection with exponential backoff
        if (reconnectAttempts.current < maxReconnectAttempts) {
          const delay = baseReconnectDelay * Math.pow(2, reconnectAttempts.current);
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttempts.current++;
            connect();
          }, delay);
        } else {
          setConnectionError('Failed to reconnect after multiple attempts');
          addNotification({
            type: 'error',
            title: 'Connection Lost',
            message: 'Lost connection to garden monitoring. Please refresh the page.',
            autoClose: false,
          });
        }
      };

      // Handle specific event types
      eventSource.addEventListener('sensor-reading', (event) => {
        const reading: SensorReading = JSON.parse(event.data);
        updateSensorReading(reading);
        
        // Invalidate related queries to trigger refetch
        queryClient.invalidateQueries({ queryKey: ['sensor-readings'] });
      });

      eventSource.addEventListener('watering-status', (event) => {
        const { zoneId, isWatering } = JSON.parse(event.data);
        setWateringStatus(zoneId, isWatering);
        
        // Show notification for watering events
        addNotification({
          type: isWatering ? 'info' : 'success',
          title: isWatering ? 'Watering Started' : 'Watering Completed',
          message: `Zone ${zoneId} ${isWatering ? 'started' : 'finished'} watering`,
        });
      });

      eventSource.addEventListener('weather-alert', (event) => {
        const alert = JSON.parse(event.data);
        addNotification({
          type: 'warning',
          title: 'Weather Alert',
          message: alert.message,
          autoClose: false,
        });
      });

      eventSource.addEventListener('system-alert', (event) => {
        const alert = JSON.parse(event.data);
        addNotification({
          type: alert.severity === 'high' ? 'error' : 'warning',
          title: 'System Alert',
          message: alert.message,
          autoClose: alert.severity !== 'high',
        });
      });

    } catch (error) {
      console.error('Failed to create EventSource:', error);
      setConnectionError('Failed to establish connection');
    }
  };

  const disconnect = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    
    setIsConnected(false);
    setConnectionError(null);
  };

  const handleRealtimeMessage = (data: any) => {
    switch (data.type) {
      case 'sensor-reading':
        updateSensorReading(data.payload);
        break;
      case 'watering-status':
        setWateringStatus(data.payload.zoneId, data.payload.isWatering);
        break;
      case 'garden-updated':
        // Invalidate garden queries
        queryClient.invalidateQueries({ queryKey: ['gardens'] });
        queryClient.invalidateQueries({ queryKey: ['garden', data.payload.gardenId] });
        break;
      default:
        console.log('Unknown real-time message type:', data.type);
    }
  };

  useEffect(() => {
    if (enabled && gardenId) {
      connect();
    }

    return disconnect;
  }, [enabled, gardenId]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, []);

  return {
    isConnected,
    connectionError,
    connect,
    disconnect,
  };
}

// Hook for WebSocket-based real-time communication
export function useWebSocketData({ enabled = true, gardenId }: UseRealTimeDataOptions = {}) {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);

  const queryClient = useQueryClient();
  const { updateSensorReading, setWateringStatus } = useGardenStore();
  const { addNotification } = useNotificationStore();

  const connect = () => {
    if (!enabled || !gardenId) return;

    try {
      const ws = apiClient.createWebSocket(`/ws/garden/${gardenId}`);
      setSocket(ws);

      ws.onopen = () => {
        console.log('WebSocket connection opened');
        setIsConnected(true);
        setConnectionError(null);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          handleWebSocketMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionError('WebSocket connection error');
      };

      ws.onclose = (event) => {
        console.log('WebSocket connection closed:', event.code, event.reason);
        setIsConnected(false);
        setSocket(null);
        
        // Attempt reconnection if not a normal closure
        if (event.code !== 1000 && enabled) {
          setTimeout(connect, 3000);
        }
      };

    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      setConnectionError('Failed to establish WebSocket connection');
    }
  };

  const disconnect = () => {
    if (socket) {
      socket.close(1000, 'Client disconnecting');
      setSocket(null);
    }
    setIsConnected(false);
    setConnectionError(null);
  };

  const sendMessage = (message: any) => {
    if (socket && isConnected) {
      socket.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  };

  const handleWebSocketMessage = (data: any) => {
    switch (data.type) {
      case 'sensor-reading':
        updateSensorReading(data.payload);
        queryClient.invalidateQueries({ queryKey: ['sensor-readings'] });
        break;
      case 'watering-status':
        setWateringStatus(data.payload.zoneId, data.payload.isWatering);
        break;
      case 'notification':
        addNotification({
          type: data.payload.type,
          title: data.payload.title,
          message: data.payload.message,
        });
        break;
      default:
        console.log('Unknown WebSocket message type:', data.type);
    }
  };

  useEffect(() => {
    if (enabled && gardenId) {
      connect();
    }

    return disconnect;
  }, [enabled, gardenId]);

  return {
    isConnected,
    connectionError,
    sendMessage,
    connect,
    disconnect,
  };
}