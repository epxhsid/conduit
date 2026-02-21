use bytes::BytesMut;

#[derive(Debug, Clone)]
pub struct Frame {
    pub cmd: u16,
    pub payload: BytesMut,
}