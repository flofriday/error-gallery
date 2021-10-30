fn hello(name: &str) -> String {
    return format!("Hi {}!", name)
}

fn main() {
    let greeting = hallo("friday");
    println!("{}", greeting);
}