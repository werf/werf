class ApplicationController < ActionController::Base
  def index
    render :json => "Hello from app 2"
  end
end
