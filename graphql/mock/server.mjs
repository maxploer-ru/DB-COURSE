import { readFile } from 'node:fs/promises';
import { createServer } from 'node:http';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { buildSchema, graphql } from 'graphql';

const here = dirname(fileURLToPath(import.meta.url));
const schemaSDL = await readFile(join(here, '..', 'schema.graphql'), 'utf8');
const schema = buildSchema(schemaSDL);
const port = Number(process.env.PORT ?? 4000);

const now = '2026-09-19T09:00:00Z';
let nextChannelId = 8;
let nextVideoId = 102;
let nextCommentId = 4;

const page = (items, args = {}) => {
  const limit = Math.max(1, Math.min(Number(args.limit ?? 20), 100));
  const offset = Math.max(0, Number(args.offset ?? 0));
  return {
    items: items.slice(offset, offset + limit),
    pageInfo: { limit, offset, totalCount: items.length },
  };
};

const makeStats = () => ({ views: 1234, likes: 87, dislikes: 4, comments: 12 });
const makeRatingStats = () => ({ likes: 87, dislikes: 4 });

const videos = [];
const channels = [];
const playlists = [];
const comments = [];
const posts = [];
const communityComments = [];

function makeComment(id, videoId, content = 'REST and GraphQL are both resource-oriented contracts.') {
  return {
    id: String(id),
    userId: '42',
    username: 'marina',
    videoId: String(videoId),
    content,
    createdAt: now,
    stats: () => makeRatingStats(),
  };
}

function makeVideo(id, channelId, title = 'REST API design') {
  const video = {
    id: String(id),
    channelId: String(channelId),
    channelName: 'databases-101',
    title,
    description: 'Designing a resource-oriented API',
    filepath: `videos/${id}/rest-api.mp4`,
    originalFilename: 'rest-api.mp4',
    status: 'READY',
    createdAt: now,
    stats: () => makeStats(),
    media: () => ({ url: 'https://storage.example.com/video/101', expiresAt: now }),
    comments: (args) => page(comments.filter((item) => item.videoId === String(id)), args),
  };
  return video;
}

function makePlaylist(id, channelId, name = 'API lessons') {
  return {
    id: String(id),
    channelId: String(channelId),
    name,
    description: 'Playlist created by the GraphQL mock',
    createdAt: now,
    videos: (args) => page([
      {
        playlistId: String(id),
        videoId: '101',
        number: 1,
        addedAt: now,
        videoTitle: 'REST API design',
        channelName: 'databases-101',
        videoStatus: 'READY',
      },
    ], args),
  };
}

function makePost(id, channelId, content = 'New community post') {
  return {
    id: String(id),
    channelId: String(channelId),
    userId: '42',
    username: 'marina',
    content,
    createdAt: now,
    comments: (args) => page(communityComments.filter((item) => item.postId === String(id)), args),
  };
}

function makeCommunity(channel, args) {
  const channelPosts = posts.filter((item) => item.channelId === channel.id);
  const result = page(channelPosts, args);
  return { channel, posts: result.items, pageInfo: result.pageInfo };
}

function makeChannel(id, name = 'databases-101') {
  const channel = {
    id: String(id),
    userId: '42',
    ownerUsername: 'marina',
    name,
    description: 'Practical database lessons',
    createdAt: now,
    videos: (args) => page(videos.filter((item) => item.channelId === String(id)), args),
    playlists: (args) => page(playlists.filter((item) => item.channelId === String(id)), args),
    community: (args) => makeCommunity(channel, args),
    subscribersCount: 128,
  };
  return channel;
}

const currentUser = {
  id: '42',
  username: 'marina',
  email: 'marina@example.com',
  isActive: true,
  notificationsEnabled: true,
  createdAt: now,
  updatedAt: now,
  role: { id: '3', name: 'USER', isDefault: true },
  channel: () => channels[0],
  videos: (args) => page(videos, args),
  subscriptions: (args) => page([
    { userId: '42', channelId: '7', channelName: 'databases-101', newVideosCount: 2, subscribedAt: now },
  ], args),
  playlists: (args) => page(playlists, args),
  community: (args) => makeCommunity(channels[0], args),
};

channels.push(makeChannel(7));
videos.push(makeVideo(101, 7));
comments.push(makeComment(1, 101));
playlists.push(makePlaylist(1, 7));
posts.push(makePost(1, 7));
communityComments.push({
  id: '1', postId: '1', userId: '42', username: 'marina',
  content: 'Welcome to the community!', createdAt: now,
});

const userPage = (args) => page([currentUser], args);
const channelPage = (items, args) => page(items, args);
const videoPage = (args) => page(videos, args);
const playlistPage = (args) => page(playlists, args);
const commentPage = (args) => page(comments, args);

const authPayload = (user = currentUser) => ({
  user,
  accessToken: 'mock-access-token',
  refreshToken: 'mock-refresh-token',
  refreshExpiresAt: '2026-10-19T09:00:00Z',
});

const findChannel = (id) => channels.find((item) => item.id === String(id)) ?? null;
const findVideo = (id) => videos.find((item) => item.id === String(id)) ?? null;
const findComment = (id) => comments.find((item) => item.id === String(id)) ?? null;
const findPlaylist = (id) => playlists.find((item) => item.id === String(id)) ?? null;

const rootValue = {
  me: () => currentUser,
  users: (args) => userPage(args),
  user: ({ id }) => (String(id) === currentUser.id ? currentUser : null),
  userChannel: ({ userId }) => (String(userId) === currentUser.id ? channels[0] : null),
  channels: ({ name, ...args }) => channelPage(name ? channels.filter((item) => item.name === name) : channels, args),
  myChannel: () => channels[0],
  channel: ({ id }) => findChannel(id),
  channelSubscriptionStatus: () => ({ subscribed: true }),
  channelSubscribersCount: () => 128,
  mySubscriptions: (args) => currentUser.subscriptions(args),
  videos: (args) => videoPage(args),
  myVideos: (args) => videoPage(args),
  channelVideos: ({ channelId, ...args }) => page(videos.filter((item) => item.channelId === String(channelId)), args),
  video: ({ id }) => findVideo(id),
  videoStats: () => makeStats(),
  videoComments: ({ videoId, ...args }) => page(comments.filter((item) => item.videoId === String(videoId)), args),
  comment: ({ id }) => findComment(id),
  commentStats: () => makeRatingStats(),
  channelPlaylists: ({ channelId, ...args }) => page(playlists.filter((item) => item.channelId === String(channelId)), args),
  myPlaylists: (args) => playlistPage(args),
  playlist: ({ id }) => findPlaylist(id),
  playlistVideos: ({ playlistId, ...args }) => findPlaylist(playlistId)?.videos(args) ?? page([], args),
  channelCommunity: ({ channelId, ...args }) => {
    const channel = findChannel(channelId);
    return channel ? makeCommunity(channel, args) : null;
  },
  myCommunity: (args) => makeCommunity(channels[0], args),
  communityPostComments: ({ postId, ...args }) => page(communityComments.filter((item) => item.postId === String(postId)), args),

  register: ({ input }) => authPayload({ ...currentUser, username: input.username, email: input.email }),
  login: () => authPayload(),
  refreshToken: () => authPayload(),
  logout: () => true,
  deleteCurrentUser: () => true,
  updateNotificationSettings: ({ input }) => {
    currentUser.notificationsEnabled = input.enabled;
    return true;
  },
  updateUserStatus: ({ input }) => ({ ...currentUser, isActive: input.status === 'ACTIVE' }),
  replaceUserRole: () => true,
  createChannel: ({ input }) => {
    const channel = makeChannel(nextChannelId++, input.name);
    channel.description = input.description ?? '';
    channels.push(channel);
    return channel;
  },
  updateChannel: ({ channelId, input }) => {
    const channel = findChannel(channelId) ?? makeChannel(channelId);
    if (input.name !== undefined) channel.name = input.name;
    if (input.description !== undefined) channel.description = input.description;
    return channel;
  },
  deleteChannel: () => true,
  subscribeToChannel: () => true,
  unsubscribeFromChannel: () => true,
  resetSubscriptionNewVideosCount: () => true,
  initializeVideoUpload: ({ channelId, input }) => {
    const video = makeVideo(nextVideoId++, channelId, input.title);
    video.status = 'PENDING';
    video.description = input.description ?? '';
    videos.push(video);
    return { video, uploadUrl: 'https://storage.example.com/presigned-upload', uploadExpiresAt: '2026-09-19T09:25:00Z' };
  },
  updateVideo: ({ videoId, input }) => ({ ...findVideo(videoId), ...input }),
  deleteVideo: () => true,
  updateVideoPublication: ({ videoId }) => ({ ...findVideo(videoId), status: 'READY' }),
  recordVideoView: () => true,
  rateVideo: () => true,
  removeVideoRating: () => true,
  addVideoComment: ({ videoId, input }) => {
    const comment = makeComment(nextCommentId++, videoId, input.content);
    comments.push(comment);
    return comment;
  },
  updateComment: ({ commentId, input }) => ({ ...findComment(commentId), content: input.content }),
  deleteComment: () => true,
  rateComment: () => true,
  removeCommentRating: () => true,
  createPlaylist: ({ channelId, input }) => {
    const playlist = makePlaylist(playlists.length + 1, channelId, input.name);
    playlist.description = input.description ?? '';
    playlists.push(playlist);
    return playlist;
  },
  updatePlaylist: ({ playlistId, input }) => ({ ...findPlaylist(playlistId), ...input }),
  deletePlaylist: () => true,
  addVideoToPlaylist: () => true,
  updatePlaylistVideoPosition: () => true,
  removeVideoFromPlaylist: () => true,
  createCommunityPost: ({ channelId, input }) => {
    const post = makePost(posts.length + 1, channelId, input.content);
    posts.push(post);
    return post;
  },
  updateCommunityPost: ({ postId, input }) => ({ ...posts.find((item) => item.id === String(postId)), content: input.content }),
  deleteCommunityPost: () => true,
  createCommunityComment: ({ postId, input }) => {
    const comment = { id: String(communityComments.length + 1), postId: String(postId), userId: '42', username: 'marina', content: input.content, createdAt: now };
    communityComments.push(comment);
    return comment;
  },
  updateCommunityComment: ({ commentId, input }) => ({ ...communityComments.find((item) => item.id === String(commentId)), content: input.content }),
  deleteCommunityComment: () => true,
};

const playground = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>ZVideo GraphQL mock</title>
<style>body{font-family:system-ui;margin:2rem;max-width:1100px}textarea{width:100%;height:260px;font:14px monospace}button{margin:1rem 0;padding:.6rem 1rem}pre{background:#f4f4f4;padding:1rem;white-space:pre-wrap}</style>
</head><body><h1>ZVideo GraphQL mock</h1>
<p>Schema: <a href="/schema.graphql">schema.graphql</a> · Endpoint: <code>/graphql</code></p>
<textarea id="query">query {
  channels(limit: 10) {
    items { id name description videos { items { id title status } } }
    pageInfo { totalCount limit offset }
  }
}</textarea><br><button id="run">Run query</button><pre id="result"></pre>
<script>document.querySelector('#run').onclick=async()=>{const query=document.querySelector('#query').value;const response=await fetch('/graphql',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({query})});document.querySelector('#result').textContent=JSON.stringify(await response.json(),null,2)};</script>
</body></html>`;

function sendJSON(res, status, body) {
  res.writeHead(status, { 'content-type': 'application/json; charset=utf-8' });
  res.end(JSON.stringify(body));
}

async function execute(query, variables, operationName) {
  return graphql({ schema, source: query, rootValue, variableValues: variables, operationName });
}

const server = createServer(async (req, res) => {
  const url = new URL(req.url, `http://${req.headers.host ?? 'localhost'}`);

  if (req.method === 'GET' && url.pathname === '/') {
    res.writeHead(302, { location: '/graphql' });
    res.end();
    return;
  }
  if (req.method === 'GET' && url.pathname === '/graphql') {
    res.writeHead(200, { 'content-type': 'text/html; charset=utf-8' });
    res.end(playground);
    return;
  }
  if (req.method === 'GET' && url.pathname === '/schema.graphql') {
    res.writeHead(200, { 'content-type': 'text/plain; charset=utf-8' });
    res.end(schemaSDL);
    return;
  }
  if (req.method === 'GET' && url.pathname === '/healthz') {
    sendJSON(res, 200, { status: 'ok' });
    return;
  }
  if (url.pathname !== '/graphql' || !['GET', 'POST'].includes(req.method)) {
    sendJSON(res, 404, { error: 'Not found' });
    return;
  }

  try {
    let payload;
    if (req.method === 'GET') {
      payload = { query: url.searchParams.get('query'), operationName: url.searchParams.get('operationName') ?? undefined };
    } else {
      const chunks = [];
      for await (const chunk of req) chunks.push(chunk);
      payload = JSON.parse(Buffer.concat(chunks).toString('utf8'));
    }
    if (typeof payload.query !== 'string' || payload.query.trim() === '') {
      sendJSON(res, 400, { errors: [{ message: 'The query field is required' }] });
      return;
    }
    const result = await execute(payload.query, payload.variables, payload.operationName);
    sendJSON(res, 200, result);
  } catch (error) {
    sendJSON(res, 400, { errors: [{ message: error instanceof Error ? error.message : String(error) }] });
  }
});

server.listen(port, '0.0.0.0', () => {
  console.log(`ZVideo GraphQL mock listening on http://localhost:${port}/graphql`);
});
