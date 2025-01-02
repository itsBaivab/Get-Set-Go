class Chat {
    constructor() {
        this.ws = null;
        this.messageQueue = [];
        this.setupWebSocket();
        this.setupEventListeners();
    }

    setupWebSocket() {
        fetch('/wsurl')
            .then(response => response.text())
            .then(wsURL => {
                this.ws = new WebSocket(wsURL);
                this.setupWebSocketHandlers();
            })
            .catch(err => console.error('WebSocket setup error:', err));
    }

    setupWebSocketHandlers() {
        this.ws.onopen = () => {
            console.log('Chat WebSocket connected');
            this.processMessageQueue();
        };

        this.ws.onmessage = (event) => {
            this.displayMessage(event.data, 'received');
        };

        this.ws.onclose = () => {
            console.log('Chat WebSocket disconnected');
            setTimeout(() => this.setupWebSocket(), 3000);
        };
    }

    setupEventListeners() {
        const messageInput = document.getElementById('message');
        const sendButton = document.getElementById('sendButton');

        if (messageInput && sendButton) {
            messageInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.sendMessage(messageInput.value);
                    messageInput.value = '';
                }
            });

            sendButton.addEventListener('click', () => {
                this.sendMessage(messageInput.value);
                messageInput.value = '';
            });
        }
    }

    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'chat',
                data: message
            }));
        }
    }

    processMessageQueue() {
        while (this.messageQueue.length > 0) {
            this.sendMessage(this.messageQueue.shift());
        }
    }

    displayMessage(message, type) {
        const chatDiv = document.getElementById('chat');
        if (chatDiv) {
            const messageDiv = document.createElement('div');
            messageDiv.className = `message ${type}`;
            messageDiv.textContent = type === 'sent' ? `You: ${message}` : `Stranger: ${message}`;
            chatDiv.appendChild(messageDiv);
            chatDiv.scrollTop = chatDiv.scrollHeight;
        }
    }
}

class VideoChat {
    constructor() {
        this.peerConnection = null;
        this.localStream = null;
        this.isConnected = false;
        this.pendingCandidates = [];
        this.init();
    }

    async init() {
        try {
            await this.setupWebRTC();
            await this.setupWebSocket();
        } catch (err) {
            console.error('Init error:', err);
        }
    }

    async setupWebRTC() {
        try {
            // Initialize peer connection first
            this.peerConnection = new RTCPeerConnection({
                iceServers: [
                    { urls: 'stun:stun.l.google.com:19302' },
                    { urls: 'stun:stun1.l.google.com:19302' }
                ]
            });

            // Set up ICE handling
            this.peerConnection.onicecandidate = (event) => {
                if (event.candidate) {
                    if (this.isConnected) {
                        this.sendMessage({
                            type: 'ice-candidate',
                            candidate: event.candidate
                        });
                    } else {
                        this.pendingCandidates.push(event.candidate);
                    }
                }
            };

            // Handle remote stream
            this.peerConnection.ontrack = (event) => {
                const remoteVideo = document.getElementById('remoteVideo');
                if (remoteVideo && event.streams[0]) {
                    remoteVideo.srcObject = event.streams[0];
                }
            };

            // Get local media stream
            this.localStream = await navigator.mediaDevices.getUserMedia({
                video: true,
                audio: true
            });

            // Add tracks to peer connection
            this.localStream.getTracks().forEach(track => 
                this.peerConnection.addTrack(track, this.localStream)
            );

            // Display local video
            const localVideo = document.getElementById('localVideo');
            if (localVideo) {
                localVideo.srcObject = this.localStream;
                try {
                    await localVideo.play();
                } catch (e) {
                    console.warn('Local video autoplay failed:', e);
                }
            }

        } catch (err) {
            console.error('WebRTC setup error:', err);
            throw err;
        }
    }

    async setupWebSocket() {
        try {
            const response = await fetch('/wsurl');
            if (!response.ok) throw new Error('Failed to get WebSocket URL');
            const wsURL = await response.text();
            
            this.ws = new WebSocket(wsURL);
            this.setupWebSocketHandlers();
        } catch (err) {
            console.error('WebSocket setup error:', err);
            setTimeout(() => this.setupWebSocket(), 3000);
        }
    }

    setupWebSocketHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket Connected');
            this.isConnected = true;
            this.sendPendingCandidates();
        };

        this.ws.onmessage = async (event) => {
            try {
                const message = JSON.parse(event.data);
                await this.handleSignalingMessage(message);
            } catch (err) {
                console.error('Message handling error:', err);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket Disconnected');
            this.isConnected = false;
            setTimeout(() => this.setupWebSocket(), 3000);
        };
    }

    async handleSignalingMessage(message) {
        try {
            switch (message.type) {
                case 'connected':
                    await this.createOffer();
                    break;

                case 'offer':
                    if (message.sdp) {
                        await this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.sdp));
                        const answer = await this.peerConnection.createAnswer();
                        await this.peerConnection.setLocalDescription(answer);
                        this.sendMessage({
                            type: 'answer',
                            sdp: answer
                        });
                    }
                    break;

                case 'answer':
                    if (message.sdp) {
                        await this.peerConnection.setRemoteDescription(new RTCSessionDescription(message.sdp));
                    }
                    break;

                case 'ice-candidate':
                    if (message.candidate) {
                        await this.peerConnection.addIceCandidate(new RTCIceCandidate(message.candidate));
                    }
                    break;
            }
        } catch (err) {
            console.error('Signaling error:', err);
        }
    }

    async createOffer() {
        try {
            if (!this.peerConnection) {
                console.error('PeerConnection not initialized');
                return;
            }
            
            const offer = await this.peerConnection.createOffer({
                offerToReceiveAudio: true,
                offerToReceiveVideo: true
            });
            
            await this.peerConnection.setLocalDescription(offer);
            
            this.sendMessage({
                type: 'offer',
                sdp: offer
            });
        } catch (err) {
            console.error('Create offer error:', err);
        }
    }

    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }

    sendPendingCandidates() {
        while (this.pendingCandidates.length > 0) {
            const candidate = this.pendingCandidates.shift();
            this.sendMessage({
                type: 'ice-candidate',
                candidate
            });
        }
    }

    cleanup() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
        }
        if (this.peerConnection) {
            this.peerConnection.close();
        }
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Initialize both Chat and VideoChat
window.addEventListener('load', () => {
    window.chat = new Chat();
    window.videoChat = new VideoChat();
    
    // Cleanup on page unload
    window.addEventListener('unload', () => {
        window.videoChat.cleanup();
    });
});